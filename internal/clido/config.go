package clido

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func GetConfig() (Config, error) {
	var config Config

	homeDir, err := os.UserHomeDir()

	if err != nil {
		return config, err
	} else {
		var path = filepath.Join(homeDir, ".config", "cli-do", "config.json")
		byteValue, _ := os.ReadFile(path)
		json.Unmarshal(byteValue, &config)

		return config, nil
	}
}
