package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

func loadConfiguration(path string) (map[string]any, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg map[string]any

	if err = yaml.UnmarshalStrict(fileBytes, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func writeConfiguration(path string, cfg map[string]any) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, b, 0600)
}
