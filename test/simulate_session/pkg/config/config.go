package config

import (
	"fmt"
	"time"

	"github.com/TrueBlocks/trueblocks-key/test/simulate_session/pkg/scenario"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	BaseUrl   string                       `toml:"baseUrl"`
	Rate      uint                         `toml:"rate"`
	Duration  time.Duration                `toml:"duration"`
	Scenarios map[string]scenario.Scenario `toml:"scenarios"`
}

var k = koanf.New(".")

func ReadConfig(filePath string) (config *Config, err error) {
	config = &Config{}

	if err = k.Load(file.Provider(filePath), toml.Parser()); err != nil {
		err = fmt.Errorf("reading file %s: %w", filePath, err)
		return
	}

	if err = k.Unmarshal("", config); err != nil {
		err = fmt.Errorf("unmarshal: %w", err)
	}
	return
}
