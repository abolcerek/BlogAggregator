package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = "gatorconfig.json"

type Config struct {
	DbUrl string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error){
	configPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, err
	}
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (cfg *Config) SetUser(username string) error {
	cfg.CurrentUserName = username
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	write(configPath, data)
	return nil
}

func getConfigFilePath() (string, error) {
	dir, dirErr := os.UserHomeDir()
	if dirErr != nil {
		return "", dirErr
	}
	configPath := filepath.Join(dir, configFileName)
	return configPath, nil
}

func write(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
