package configs

import (
	"fmt"
	"os"

	_ "embed"

	"gopkg.in/yaml.v3"
)

//go:embed local.yaml
var DefaultConfigBytes []byte

type Config struct {
	Database    Database    `yaml:"database"`
	Http        HTTP        `yaml:"http"`
	Logic       Logic       `yaml:"logic"`
	TestCaseRun TestCaseRun `yaml:"test_case_run"`
	Token       Token       `yaml:"token"`
}

func NewConfig(filePath string) (Config, error) {
	var (
		configBytes = DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(filePath)
		if err != nil {
			return Config{}, fmt.Errorf("failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return config, nil
}
