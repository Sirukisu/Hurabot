package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type MainBotConfig struct {
	AuthenticationToken string
	CommandPrefix       string
	GuildID             string
	ModelToUse          string
	LogDir              string
	LogLevel            string
}

func LoadConfig(configFile string) MainBotConfig {
	runDirectory, err := os.Executable()

	if err != nil {
		log.Fatal("Unable to find run directory: " + err.Error())
	}

	if configFile == "" {
		// assume config is in default directory
		configFile = filepath.Join(runDirectory, "config.json")
	}

	// check if config file exists & read
	configFileContent, err := os.ReadFile(configFile)

	if os.IsNotExist(err) {
		// config file not found, make a new one
		log.Println("Config file " + configFile + " not found, creating a new one")
		BotConfig := CreateEmptyConfig(configFile)
		return BotConfig
	} else if err != nil {
		log.Fatal("Error reading file " + configFile + ": " + err.Error())
	}

	var botConfig MainBotConfig
	err = json.Unmarshal(configFileContent, &botConfig)

	if err != nil {
		log.Fatal("Error reading config " + configFile + ": " + err.Error())
	}

	return botConfig
}

func CreateEmptyConfig(configFile string) MainBotConfig {
	botConfig := MainBotConfig{
		AuthenticationToken: "",
		CommandPrefix:       "",
		GuildID:             "",
		ModelToUse:          "",
		LogDir:              "/log/",
		LogLevel:            "default",
	}

	botConfigJson, err := json.Marshal(botConfig)

	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(configFile, botConfigJson, 0664)

	if err != nil {
		log.Println("Failed to write new config to " + configFile + ": " + err.Error())
	}

	return botConfig
}
