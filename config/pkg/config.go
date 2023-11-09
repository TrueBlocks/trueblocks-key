package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const prefix = "QNEXT"

type ConfigFile struct {
	Version     string
	Chains      chainsGroup
	Database    map[string]databaseGroup
	Sqs         sqsGroup
	Query       queryGroup
	QnProvision qnProvisionGroup `koanf:"qnprovision"`
}

type chainsGroup struct {
	// map of chain name to network names,
	// e.g. "ethereum" => ["mainnet"]
	Allowed map[string][]string
}

type databaseGroup struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	AwsSecret string
}

type sqsGroup struct {
	QueueName       string
	InsertBatchSize uint
}

type queryGroup struct {
	MaxLimit uint
}

type qnProvisionGroup struct {
	TableName    string
	AuthUsername string
	AuthPassword string
	AwsSecret    string
}

var cached *ConfigFile
var k = koanf.New(".")

func Get(configPath string) (*ConfigFile, error) {
	if cached != nil {
		return cached, nil
	}

	// Load defaults
	err := k.Load(confmap.Provider(defaultConfig, "."), nil)
	if err != nil {
		return nil, fmt.Errorf("config: setting default values: %w", err)
	}

	if configPath != "" {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return nil, fmt.Errorf("config: reading file %s: %w", configPath, err)
		}
	}

	translateEnv := func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "QNEXT_")), "_", ".", -1)
	}
	if err := k.Load(env.Provider(prefix+"_", ".", translateEnv), nil); err != nil {
		return nil, fmt.Errorf("config: parsing env: %w", err)
	}

	var out ConfigFile
	if err := k.Unmarshal("", &out); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	cached = &out
	return cached, nil
}
