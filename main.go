package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"os"
)

func main() {
	parser := argparse.NewParser("Hurabot", "A Discord bot that uses Markov chains to generate random text from your messages")

	// MODEL COMMANDS
	modelCommand := parser.NewCommand("model", "manage bot word models")

	// model creation command
	modelCommandCreate := modelCommand.NewCommand("create", "create new model from discord messages")
	modelCommandCreateArgs := modelCommandCreate.File("f", "file", os.O_RDONLY, 0660, &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Discord messages folder to process",
		Default:  nil,
	})
	modelCommandList := modelCommand.NewCommand("list", "List current models")

	// model text generation command
	modelCommandGenerate := modelCommand.NewCommand("generate", "Generate random text from a model")
	modelCommandModelFileArg := modelCommandGenerate.File("m", "model", os.O_RDONLY, 0440, &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Model file to use",
		Default:  nil,
	})
	modelCommandGenerateCountArg := modelCommandGenerate.Int("w", "words", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Amount of words to generate",
		Default:  10,
	})

	// TODO model delete command, is it necessary?
	modelCommandRemove := modelCommand.NewCommand("remove", "Remove a model")

	// CONFIG OPTIONS
	configCommand := parser.NewCommand("config", "config options")

	// config file options
	configCommandFileOptions := &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Config file to use",
		Default:  "config.json",
	}

	configCommandConfigFile := configCommand.File("c", "config-file", os.O_RDWR|os.O_CREATE, 0660, configCommandFileOptions)

	configCommandShow := configCommand.NewCommand("show", "Show config")
	configCommandCreate := configCommand.NewCommand("create", "Create a blank config")
	configCommandEdit := configCommand.NewCommand("edit", "Edit config file")

	// BOT COMMANDS
	botCommand := parser.NewCommand("bot", "bot options")

	botCommandRun := botCommand.NewCommand("run", "Run the bot")
	botCommandRunConfigArg := botCommandRun.File("c", "config-file", os.O_RDONLY, 0660, configCommandFileOptions)

	// END OF ARGUMENTS

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		fmt.Println("Type -h for usage info")
		return
	}

	// handle model commands
	if modelCommandCreate.Happened() {
		CreateModelInit(modelCommandCreateArgs)
	}

	if modelCommandList.Happened() {
		//TODO model command list
	}

	if modelCommandRemove.Happened() {
		//TODO model command remove
	}

	if modelCommandGenerate.Happened() {
		wordModel, err := LoadModel(modelCommandModelFileArg)

		if err != nil {
			fmt.Println("Failed to load model " + modelCommandModelFileArg.Name())
			return
		}
		fmt.Printf("Loaded %d words from model %s\n", len(wordModel.Words), wordModel.Name)

		generatedText := GenerateWords(wordModel, modelCommandGenerateCountArg)
		fmt.Println(generatedText)
	}

	// handle config commands
	if configCommandShow.Happened() {
		ConfigShowConfig(configCommandConfigFile)
	}
	if configCommandCreate.Happened() {
		ConfigCreateEmptyConfig(configCommandConfigFile)
	}
	if configCommandEdit.Happened() {
		ConfigEdit(configCommandConfigFile)
	}

	// handle bot commands
	if botCommandRun.Happened() {
		if err := ConfigLoadConfig(botCommandRunConfigArg); err != nil {
			fmt.Printf("Failed to load config from %s: %s", botCommandRunConfigArg.Name(), err.Error())
			return
		}
		if err := RunBot(); err != nil {
			fmt.Println(err)
			return
		}
	}
}
