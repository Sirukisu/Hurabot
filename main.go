package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
	"strings"
)

func main() {
	parser := argparse.NewParser("Hurabotti", "Botin tynkä")

	// model commands
	modelCommand := parser.NewCommand("model", "manage bot word models")
	modelCommandCreate := modelCommand.NewCommand("create", "create new model")
	modelCommandCreateArgs := modelCommandCreate.StringList("f", "file", &argparse.Options{
		Required: true,
		Validate: verifyCsv,
		Help:     "CSV file or folder to process",
		Default:  nil,
	})
	modelCommandList := modelCommand.NewCommand("list", "List current models")
	modelCommandRemove := modelCommand.NewCommand("remove", "Remove a model")

	// bot commands
	botCommand := parser.NewCommand("bot", "bot options")

	// bot config options
	botCommandConfigOptions := &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Config file to use",
		Default:  "config.json",
	}

	botCommandConfig := botCommand.NewCommand("config", "Manage bot config")
	botCommandConfigFile := botCommandConfig.File("f", "config-file", os.O_RDWR|os.O_CREATE, 0660, botCommandConfigOptions)

	botCommandConfigShow := botCommandConfig.NewCommand("show", "Show config")
	botCommandConfigEdit := botCommandConfig.NewCommand("edit", "Edit config file")
	//botCommandConfigEditFile := botCommandConfigEdit.File("f", "config-file", os.O_RDWR, 0660, botCommandConfigOptions)

	//botCommandRun := botCommand.NewCommand("run", "Run the bot")

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		fmt.Println("Type -h for usage info")
	}

	// handle model commands
	if modelCommandCreate.Happened() {
		CreateModel(modelCommandCreateArgs)
	} else if modelCommandList.Happened() {

	} else if modelCommandRemove.Happened() {

	}

	// handle bot config commands
	if botCommandConfigShow.Happened() {
		BotShowConfig(botCommandConfigFile)
	} else if botCommandConfigEdit.Happened() {
		EditConfig(LoadConfig(botCommandConfigFile))
	}
}

func verifyCsv(args []string) error {
	fileInfo, err := os.Stat(args[0])

	if err != nil {
		//log.Fatal(err)
		return nil
	}

	if fileInfo.IsDir() == false {
		if strings.HasSuffix(args[0], ".csv") == false {
			//log.Fatal("File doesn't end with .csv.")
		}
	}
	return nil
}
