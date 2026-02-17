package app

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func FindProfile(cfg Config, name string) (*Profile, bool) {
	for i := range cfg.Profiles {
		if cfg.Profiles[i].Name == name {
			return &cfg.Profiles[i], true
		}
	}

	return nil, false
}
