package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Settings holds app settings. Stored in config dir.
type Settings struct {
	// CardsDir is the directory containing PNG character cards.
	CardsDir string `json:"cards_dir"`
}

// SettingsPath returns the path to settings.json in the config directory.
func SettingsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "settings.json"), nil
}

// GetSettings loads settings from the store. Missing file or empty CardsDir is not an error.
func GetSettings() (Settings, error) {
	path, err := SettingsPath()
	if err != nil {
		return Settings{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Settings{}, nil
		}
		return Settings{}, fmt.Errorf("read settings: %w", err)
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, fmt.Errorf("decode settings: %w", err)
	}
	return s, nil
}

// SetSettings saves settings. Creates config dir if needed.
func SetSettings(s Settings) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path, err := SettingsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}
