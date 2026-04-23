package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the persisted CLI settings for one user.
type Config struct {
	BackendURL   string  `json:"backend_url"`
	AuthToken    string  `json:"auth_token"`
	RefreshToken *string `json:"refresh_token,omitempty"`
	CsrfToken    *string `json:"csrf_token,omitempty"`
}

// IsComplete returns true when both URL and auth token are present.
func (c *Config) IsComplete() bool {
	return c.BackendURL != "" && c.AuthToken != ""
}

// Path returns the resolved config file path, honouring an optional override.
func Path(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "alvestass-admin", "config.json"), nil
}

// Load reads the config file, returning an empty Config when the file does not
// exist. A parse error returns an empty Config so the wizard can re-run.
func Load(override string) (*Config, error) {
	p, err := Path(override)
	if err != nil {
		return &Config{}, nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &Config{}, nil
	}
	return &cfg, nil
}

// Save writes the config to disk with 0600 permissions.
func (c *Config) Save(override string) error {
	p, err := Path(override)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}
