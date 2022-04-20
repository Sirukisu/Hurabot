package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type MainBotConfig struct {
	AuthenticationToken string
	CommandPrefix       string
	GuildID             string
	ModelToUse          string
	LogDir              string
	LogLevel            string
}

func LoadConfig(configFile *os.File) MainBotConfig {
	var botConfig MainBotConfig
	configFileContent, err := os.ReadFile(configFile.Name())

	if err != nil {
		log.Fatal("Error reading file " + configFile.Name() + ": " + err.Error())
	}

	// check if file is empty = just created, make new config
	if len(configFileContent) == 0 {
		botConfig = CreateEmptyConfig(configFile)
		return botConfig
	}

	err = json.Unmarshal(configFileContent, &botConfig)

	if err != nil {
		log.Fatal("Error reading parameters from " + configFile.Name() + ": " + err.Error())
	}

	return botConfig
}

func CreateEmptyConfig(configFile *os.File) MainBotConfig {
	botConfig := MainBotConfig{
		AuthenticationToken: "",
		CommandPrefix:       "",
		GuildID:             "",
		ModelToUse:          "",
		LogDir:              "/log/",
		LogLevel:            "default",
	}

	botConfigJson, err := json.MarshalIndent(botConfig, "", "\t")

	if err != nil {
		log.Fatal(err)
	}

	_, err = configFile.Write(botConfigJson)

	if err != nil {
		log.Println("Failed to write new config to " + configFile.Name() + ": " + err.Error())
	}

	fmt.Println("Wrote a new config to " + configFile.Name())
	return botConfig
}
