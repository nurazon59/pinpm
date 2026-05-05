package pinpm

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Version int      `yaml:"version"`
	Ignore  []string `yaml:"ignore,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{Version: 1}
	if path == "" {
		return cfg, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}

	err = yaml.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func MarshalResultJSON(result *Result) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
