package api

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type GlobalSettings struct {
	MaxGlobalContainers      int `json:"max_global_containers" yaml:"max_global_containers"`
	MaxGlobalBuildContainers int `json:"max_global_build_containers" yaml:"max_global_build_containers"`
	MaxGlobalChatContainers  int `json:"max_global_chat_containers" yaml:"max_global_chat_containers"`
	MaxGlobalCmdContainers   int `json:"max_global_cmd_containers" yaml:"max_global_cmd_containers"`
}

func GetGlobalSettingsPath() string {
	return filepath.Join("projects", "global_settings.yml")
}

func LoadGlobalSettings() (*GlobalSettings, error) {
	path := GetGlobalSettingsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults
			return &GlobalSettings{
				MaxGlobalContainers:      10,
				MaxGlobalBuildContainers: 5,
				MaxGlobalChatContainers:  5,
				MaxGlobalCmdContainers:   5,
			}, nil
		}
		return nil, err
	}

	var settings GlobalSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	// Update global variables in worker.go
	MaxGlobalContainers = settings.MaxGlobalContainers
	MaxGlobalBuildContainers = settings.MaxGlobalBuildContainers
	MaxGlobalChatContainers = settings.MaxGlobalChatContainers
	MaxGlobalCmdContainers = settings.MaxGlobalCmdContainers

	return &settings, nil
}

func SaveGlobalSettings(settings *GlobalSettings) error {
	path := GetGlobalSettingsPath()
	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Update global variables in worker.go
	MaxGlobalContainers = settings.MaxGlobalContainers
	MaxGlobalBuildContainers = settings.MaxGlobalBuildContainers
	MaxGlobalChatContainers = settings.MaxGlobalChatContainers
	MaxGlobalCmdContainers = settings.MaxGlobalCmdContainers

	return nil
}
