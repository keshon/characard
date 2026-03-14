package config

import (
	"os"
	"path/filepath"
	"strings"
)

// IsPortable returns true when CHARCARD_PORTABLE is set to 1, true, or yes (case-insensitive).
func IsPortable() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("CHARCARD_PORTABLE")))
	return v == "1" || v == "true" || v == "yes"
}

func exeDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exe), nil
}

func userConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "charcard"), nil
}

// ConfigDir returns the directory for config files (portable: next to binary; otherwise OS config dir).
func ConfigDir() (string, error) {
	if IsPortable() {
		return exeDir()
	}
	return userConfigDir()
}
