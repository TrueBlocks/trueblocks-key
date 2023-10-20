package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const prefix = "QNEXT"

type ConfigFile struct {
	Version     string
	Database    map[string]databaseGroup
	Sqs         sqsGroup
	Query       queryGroup
	QnProvision qnProvisionGroup
	QnPlans     map[string]qnPlanGroup
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
	// ARN of the API that the authorizer grants access to
	// ApiArn string -- comes in an event from api gateway
}

type qnPlanGroup struct {
	// Plan slug (name) that is registered in QN marketplace
	QnSlug string
	// Api Gateway (AWS) key id
	AwsApiKey string
}

var cached *ConfigFile
var k = koanf.New(".")

func Get(configPath string) (*ConfigFile, error) {
	if cached != nil {
		return cached, nil
	}

	if configPath != "" {
		if err := k.Load(file.Provider(configPath), toml.Parser()); err != nil {
			return nil, fmt.Errorf("config.extract: reading file %s: %w", configPath, err)
		}
	}

	translateEnv := func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, "QNEXT_")), "_", ".", -1)
	}
	if err := k.Load(env.Provider(prefix+"_", ".", translateEnv), nil); err != nil {
		return nil, fmt.Errorf("config.extract: parsing env: %w", err)
	}

	var out ConfigFile
	if err := k.Unmarshal("", &out); err != nil {
		return nil, fmt.Errorf("config.extract: unmarshal: %w", err)
	}
	cached = &out
	return cached, nil
}
