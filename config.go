package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"os"
	"strings"
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

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("New config created, edit it? (y/n)")
	resultChar, err := reader.ReadString('\n')

	if err != nil {
		log.Fatal(err)
	}
	resultChar = strings.ToLower(resultChar)

	switch resultChar {
	case "y":
		botConfig = EditConfig(botConfig)
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
	return botConfig
}

func EditConfig(config MainBotConfig) MainBotConfig {
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
}
