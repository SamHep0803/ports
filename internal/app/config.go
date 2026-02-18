package app

import (
	"fmt"
	"os"
	"strings"

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
	if err := validateConfig(cfg); err != nil {
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

func validateConfig(cfg Config) error {
	seen := make(map[string]struct{}, len(cfg.Profiles))
	for i, p := range cfg.Profiles {
		profilePath := fmt.Sprintf("profiles[%d]", i)

		if strings.TrimSpace(p.Name) == "" {
			return fmt.Errorf("%s.name is required", profilePath)
		}
		if _, ok := seen[p.Name]; ok {
			return fmt.Errorf("duplicate profile name %q", p.Name)
		}
		seen[p.Name] = struct{}{}

		if strings.TrimSpace(p.Host) == "" {
			return fmt.Errorf("%s.host is required", profilePath)
		}
		if strings.TrimSpace(p.User) == "" {
			return fmt.Errorf("%s.user is required", profilePath)
		}
		if strings.TrimSpace(p.KeyPath) == "" {
			return fmt.Errorf("%s.keyPath is required", profilePath)
		}

		for j, f := range p.Forwards {
			forwardPath := fmt.Sprintf("%s.forwards[%d]", profilePath, j)
			if strings.TrimSpace(f.RemoteHost) == "" {
				return fmt.Errorf("%s.remoteHost is required", forwardPath)
			}
			if f.LocalPort < 1 || f.LocalPort > 65535 {
				return fmt.Errorf("%s.localPort must be between 1 and 65535", forwardPath)
			}
			if f.RemotePort < 1 || f.RemotePort > 65535 {
				return fmt.Errorf("%s.remotePort must be between 1 and 65535", forwardPath)
			}
		}
	}

	return nil
}
