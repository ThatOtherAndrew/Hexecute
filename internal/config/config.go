package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
			log.Printf("Creating default settings file at %s", settingsPath)
			if err := createDefaultSettings(settingsPath, defaultSettings); err != nil {
				log.Printf("Failed to create default settings file: %v", err)
			}
			return defaultSettings, nil
		}
		return nil, err
	}

	// Check for unrecognised keys
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(data, &rawSettings); err != nil {
		log.Printf("Invalid settings file, using defaults: %v", err)
		return defaultSettings, nil
	}

	knownKeys := getKnownKeys(Settings{})
	for key := range rawSettings {
		if !knownKeys[key] {
			log.Printf("Warning: unrecognised setting key '%s' in settings file", key)
		}
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

func createDefaultSettings(path string, settings *Settings) error {
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func getKnownKeys(v interface{}) map[string]bool {
	keys := make(map[string]bool)
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			// Handle json tags like "field,omitempty"
			tagName := strings.Split(jsonTag, ",")[0]
			if tagName != "-" {
				keys[tagName] = true
			}
		}
	}
	return keys
}
