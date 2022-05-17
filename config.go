package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
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
	// Maximum amount of words that can be generated with the Discord bot
	MaxWords int
	// Logging directory
	LogDir string
	// Logging level
	LogLevel string
}

func (config MainBotConfig) createNewConfig() MainBotConfig {
	ed, err := os.Executable()

	if err != nil {
		log.Fatalln(err)
	}

	config.AuthenticationToken = ""
	config.GuildID = ""
	config.ModelDirectory = path.Join(path.Dir(ed), "models")
	config.ModelsToUse = make([]string, 0)
	config.MaxWords = 200
	config.LogDir = path.Join(path.Dir(ed), "logs")
	config.LogLevel = "default"

	return config
}

// LoadedConfig the current loaded configuration
var LoadedConfig *MainBotConfig

// ConfigLoadConfig loads the MainBotConfig from an os.File
func ConfigLoadConfig(configFile *os.File) error {

	// try to load from default location if config is not provided
	if configFile == nil {
		ed, err := os.Executable()
		if err != nil {
			return fmt.Errorf("error finding executable directory: %v", err)
		}

		defaultConfigFile, err := os.Open(path.Join(path.Dir(ed), "config.json"))
		if err != nil {
			return fmt.Errorf("no config found in %s: %v", ed, err)
		}
		configFile = defaultConfigFile
	}

	configFileContent, err := os.ReadFile(configFile.Name())

	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(configFileContent, &LoadedConfig); err != nil {
		return fmt.Errorf("failed to decode config file: %v", err)
	}

	return nil
}

// ConfigShowConfig loads and displays the config settings from an os.File
func ConfigShowConfig(configFile *os.File) error {

	// load the config
	if err := ConfigLoadConfig(configFile); err != nil {
		return fmt.Errorf("failed to load config from %s: %s", configFile.Name(), err.Error())
	}

	// print info
	fmt.Printf("Config file: %s\n"+
		"Discord authentication token: %s\n"+
		"Discord guild ID: %s\n"+
		"Models directory: %s\n"+
		"Models to use: (%d total)\n",
		configFile.Name(), LoadedConfig.AuthenticationToken, LoadedConfig.GuildID, LoadedConfig.ModelDirectory,
		len(LoadedConfig.ModelsToUse))

	for i := range LoadedConfig.ModelsToUse {
		fmt.Println(LoadedConfig.ModelsToUse[i])
	}
	fmt.Printf("Maximum words: %d\n"+
		"Log directory: %s\n"+
		"Logging level: %s",
		LoadedConfig.MaxWords, LoadedConfig.LogDir, LoadedConfig.LogLevel)

	return nil
}

// ConfigCreateEmptyConfig creates a new MainBotConfig and writes it to an os.File
func ConfigCreateEmptyConfig(configFile *os.File) error {

	botConfig := MainBotConfig{}
	botConfig = botConfig.createNewConfig()
	LoadedConfig = &botConfig

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("New config created, edit it? (y/n)")
	resultChar, _, err := reader.ReadRune()

	if err != nil {
		return err
	}
	if strings.ToLower(string(resultChar)) == "y" {
		ConfigEditCUI(nil)
	}

	configFileInfo, err := os.Stat(configFile.Name())

	if !os.IsNotExist(err) && err != nil {
		return fmt.Errorf("failed to stat config file %s: %v", configFile.Name(), err)
	} else if os.IsNotExist(err) {
		if err := configFile.Truncate(0); err != nil {
			fmt.Printf("Failed to truncate config file %s: %v\n", configFile.Name(), err)
		}
	} else if configFileInfo.Size() != 0 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Config file %s already exists, overwrite? (y/n)", configFile.Name())
		resultChar, _, err := reader.ReadRune()

		if err != nil {
			return err
		}

		switch strings.ToLower(string(resultChar)) {
		case "y":
			if err := configFile.Truncate(0); err != nil {
				fmt.Printf("Failed to truncate config file %s: %v\n", configFile.Name(), err)
			}
		case "n":
			return fmt.Errorf("aborted by user")
		default:
			return fmt.Errorf("invalid input, only use y/n")
		}
	}

	enc := json.NewEncoder(configFile)
	enc.SetIndent("", "\t")
	if err := enc.Encode(botConfig); err != nil {
		return fmt.Errorf("failed to write config to %s: %v", configFile.Name(), err.Error())
	}

	fmt.Println("Wrote a new config to " + configFile.Name())

	if err := configFile.Close(); err != nil {
		fmt.Printf("Failed to close config file %s: %v\n", configFile.Name(), err)
	}

	return nil
}

func ConfigEdit(configFile *os.File) {

	ConfigEditCUI(configFile)

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

	if err := configFile.Close(); err != nil {
		log.Printf("Failed to close config file %s: %v", configFile.Name(), err)
	}

	fmt.Println("Edited config at " + configFile.Name())
}
