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
	ModelFolder string
	// List of models to use if we don't want everything in ModelFolder
	ModelsToUse []string
	// Logging directory
	LogDir string
	// Logging level
	LogLevel string
}

// LoadConfig loads the MainBotConfig from an os.File
func LoadConfig(configFile *os.File) *MainBotConfig {
	var botConfig *MainBotConfig
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

// CreateEmptyConfig creates a new MainBotConfig and writes it to an os.File
func CreateEmptyConfig(configFile *os.File) *MainBotConfig {
	wd, err := os.Getwd()

	if err != nil {
		fmt.Println(err)
	}

	botConfig := MainBotConfig{
		AuthenticationToken: "",
		GuildID:             "",
		ModelFolder:         wd + string(os.PathSeparator) + "models" + string(os.PathSeparator),
		ModelsToUse:         make([]string, 0),
		LogDir:              wd + string(os.PathSeparator) + "logs" + string(os.PathSeparator),
		LogLevel:            "default",
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("New config created, edit it? (y/n)")
	resultChar, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}
	resultChar = strings.ToLower(resultChar)

	switch resultChar {
	case "y":
		//botConfig = EditConfig(botConfig)
	}

	botConfigJson, err := json.MarshalIndent(botConfig, "", "\t")

	if err != nil {
		log.Fatal(err)
	}

	_, err = configFile.Write(botConfigJson)

	if err != nil {
		log.Fatal("Failed to write new config to " + configFile.Name() + ": " + err.Error())
	}

	fmt.Println("Wrote a new config to " + configFile.Name())
	return &botConfig
}

/*func EditConfig(config MainBotConfig) MainBotConfig {
	gui, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		log.Panicln(err)
	}

	defer gui.Close()

	gui.SetManagerFunc(EditConfigLayout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	return config
}

func EditConfigLayout(gui *gocui.Gui) error {
	maxX, maxY := gui.Size()

	if v, err := gui.SetView("Edit config", maxX/3, maxY/3, maxX/3, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "gui test")
	}

	return nil
}

func quit(gui *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}*/
