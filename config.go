package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// MainBotConfig contains configuration of the bot
type MainBotConfig struct {
	// Discord authentication token
	AuthenticationToken string
	// Guild ID to register commands
	GuildID string
	// Folder where models are stored
	ModelDirectory string
	// List of models to use if we don't want everything in ModelFolder
	ModelsToUse []string
	// Logging directory
	LogDir string
	// Logging level
	LogLevel string
}

func (config MainBotConfig) createNewConfig() MainBotConfig {
	wd, err := os.Getwd()

	if err != nil {
		log.Fatalln(err)
	}

	config.AuthenticationToken = ""
	config.GuildID = ""
	config.ModelDirectory = wd + string(os.PathSeparator) + "models" + string(os.PathSeparator)
	config.ModelsToUse = make([]string, 0)
	config.LogDir = wd + string(os.PathSeparator) + "logs" + string(os.PathSeparator)
	config.LogLevel = "default"

	return config
}

var LoadedConfig *MainBotConfig

// ConfigLoadConfig loads the MainBotConfig from an os.File
func ConfigLoadConfig(configFile *os.File) error {
	//var botConfig *MainBotConfig
	configFileContent, err := os.ReadFile(configFile.Name())

	if err != nil {
		return err
	}

	// check if file is empty = just created, make new config
	if len(configFileContent) == 0 {
		ConfigCreateEmptyConfig(configFile)
		os.Exit(0)
	}

	err = json.Unmarshal(configFileContent, &LoadedConfig)

	if err != nil {
		return err
	}
	return nil
}

// ConfigShowConfig loads and displays the config settings from an os.File
func ConfigShowConfig(configFile *os.File) {
	// load the config
	if err := ConfigLoadConfig(configFile); err != nil {
		log.Fatalf("Failed to load config from %s: %s", configFile.Name(), err.Error())
	}

	// print info
	fmt.Println("Config file: " + configFile.Name())
	fmt.Println("Discord authentication token: " + LoadedConfig.AuthenticationToken)
	fmt.Println("Discord guild ID: " + LoadedConfig.GuildID)
	fmt.Println("Models directory: " + LoadedConfig.ModelDirectory)
	fmt.Printf("Models to use: (%d total)\n", len(LoadedConfig.ModelsToUse))
	for i := range LoadedConfig.ModelsToUse {
		fmt.Println(LoadedConfig.ModelsToUse[i])
	}
	fmt.Println("Log directory: " + LoadedConfig.LogDir)
	fmt.Println("Logging level: " + LoadedConfig.LogLevel)
}

// ConfigCreateEmptyConfig creates a new MainBotConfig and writes it to an os.File
func ConfigCreateEmptyConfig(configFile *os.File) {
	botConfig := MainBotConfig{}
	botConfig = botConfig.createNewConfig()
	LoadedConfig = &botConfig

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("New config created, edit it? (y/n)")
	resultChar, _, err := reader.ReadRune()

	if err != nil {
		log.Fatalln(err)
	}
	if strings.ToLower(string(resultChar)) == "y" {
		ConfigEditGUI(nil)
	}

	botConfigJson, err := json.MarshalIndent(botConfig, "", "\t")

	if err != nil {
		log.Fatalln(err)
	}

	_, err = configFile.Write(botConfigJson)

	if err != nil {
		log.Fatalln("Failed to write new config to " + configFile.Name() + ": " + err.Error())
	}

	fmt.Println("Wrote a new config to " + configFile.Name())
}

func ConfigEdit(configFile *os.File) {

	ConfigEditGUI(configFile)

	botConfigJson, err := json.MarshalIndent(LoadedConfig, "", "\t")

	if err != nil {
		log.Fatalln(err)
	}

	err = configFile.Truncate(0)

	if err != nil {
		log.Fatalln(err)
	}

	_, err = configFile.Write(botConfigJson)

	if err != nil {
		log.Fatalln("Failed to write new config to " + configFile.Name() + ": " + err.Error())
	}

	fmt.Println("Edited config at " + configFile.Name())
}
