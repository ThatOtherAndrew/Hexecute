package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// TODO: migrate other settings
type Settings struct {
	OverlayAlpha float32 `json:"overlay_alpha"`
}

func GetPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config", "hexecute")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "gestures.json"), nil
}

func GetSettingsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(homeDir, ".config", "hexecute")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(configDir, "settings.json"), nil
}

func LoadSettings() (*Settings, error) {
	settingsPath, err := GetSettingsPath()
	if err != nil {
		return nil, err
	}

	defaultSettings := &Settings{
		OverlayAlpha: 0.75,
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultSettings, nil
		}
		return nil, err
	}

	settings := &Settings{}
	if err := json.Unmarshal(data, settings); err != nil {
		return defaultSettings, err
	}

	return settings, nil
}
