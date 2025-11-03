package config

import (
	"encoding/json"
	"log"
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
		log.Printf("Invalid settings file, using defaults: %v", err)
		return defaultSettings, nil
	}

	// Validate and clamp overlay_alpha to [0, 1]
	if settings.OverlayAlpha < 0.0 || settings.OverlayAlpha > 1.0 {
		log.Printf("Invalid overlay_alpha value %.2f, must be between 0.0 and 1.0, using default %.2f",
			settings.OverlayAlpha, defaultSettings.OverlayAlpha)
		settings.OverlayAlpha = defaultSettings.OverlayAlpha
	}

	return settings, nil
}
