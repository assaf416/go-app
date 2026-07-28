package config

import (
	"encoding/json"
	"os"
)

// AS400Config holds the connection details for the comtec-as400 integration,
// persisted by `go-app --setup`.
type AS400Config struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func SaveAS400(path string, cfg *AS400Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadAS400(path string) (*AS400Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AS400Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
