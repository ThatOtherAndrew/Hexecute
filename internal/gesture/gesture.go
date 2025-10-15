package gestures

import (
	"encoding/json"
	"os"

	"github.com/ThatOtherAndrew/Hexecute/internal/config"
	"github.com/ThatOtherAndrew/Hexecute/internal/models"
)

func LoadGestures() ([]models.GestureConfig, error) {
	configFile, err := config.GetPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.GestureConfig{}, nil
		}
		return nil, err
	}

	var gestures []models.GestureConfig
	if err := json.Unmarshal(data, &gestures); err != nil {
		return nil, err
	}

	return gestures, nil
}

func SaveGesture(command string, templates [][]models.Point) error {
	configFile, err := config.GetPath()
	if err != nil {
		return err
	}

	var gestures []models.GestureConfig
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &gestures)
	}

	newGesture := models.GestureConfig{
		Command:   command,
		Templates: templates,
	}

	found := false
	for i, g := range gestures {
		if g.Command == command {
			gestures[i] = newGesture
			found = true
			break
		}
	}
	if !found {
		gestures = append(gestures, newGesture)
	}

	data, err := json.Marshal(gestures)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}
