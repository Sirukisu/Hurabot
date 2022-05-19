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

	modelCommandModelFileOptions := &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Model file to use",
		Default:  nil,
	}

	// model creation command
	modelCommandCreate := modelCommand.NewCommand("create", "create new model from discord messages")
	modelCommandCreateArgs := modelCommandCreate.File("d", "directory", os.O_RDONLY, 0660, &argparse.Options{
		Required: true,
		Validate: nil,
		Help:     "Discord messages folder to process",
		Default:  nil,
	})

	// model show command
	modelCommandShow := modelCommand.NewCommand("show", "show info from a model")
	modelCommandShowArgs := modelCommandShow.FileList("m", "model", os.O_RDONLY, 0440, modelCommandModelFileOptions)

	// model text generation command
	modelCommandGenerate := modelCommand.NewCommand("generate", "Generate random text from a model")
	modelCommandModelFileArg := modelCommandGenerate.File("m", "model", os.O_RDONLY, 0440, modelCommandModelFileOptions)
	modelCommandGenerateCountArg := modelCommandGenerate.Int("w", "words", &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Amount of words to generate",
		Default:  10,
	})

	// CONFIG OPTIONS
	configCommand := parser.NewCommand("config", "config options")

	// config file options
	configCommandFileOptions := &argparse.Options{
		Required: false,
		Validate: nil,
		Help:     "Config file to use",
		Default:  "config.json",
	}

	configCommandShow := configCommand.NewCommand("show", "Show config")
	configCommandCreate := configCommand.NewCommand("create", "Create a blank config")
	configCommandEdit := configCommand.NewCommand("edit", "Edit config file")

	configCommandShowConfigFile := configCommandShow.File("c", "config-file", os.O_RDONLY, 0660, configCommandFileOptions)
	configCommandCreateConfigFile := configCommandCreate.File("c", "config-file", os.O_RDWR|os.O_CREATE, 0660, configCommandFileOptions)
	configCommandEditConfigFile := configCommandEdit.File("c", "config-file", os.O_RDWR, 0660, configCommandFileOptions)

	// BOT COMMANDS
	botCommand := parser.NewCommand("run", "run the bot")
	botCommandRunConfigArg := botCommand.File("c", "config-file", os.O_RDONLY, 0660, configCommandFileOptions)

	// END OF ARGUMENTS

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		fmt.Println("Type -h for usage info")
		return
	}
	// handle model commands
	if modelCommandCreate.Happened() {
		if err := CreateModel(modelCommandCreateArgs); err != nil {
			fmt.Printf("Error creating model: %v", err)
		}
		return
	}
	if modelCommandShow.Happened() {
		if len(*modelCommandShowArgs) == 0 {
			fmt.Println("No models provided")
			return
		}

		for _, file := range *modelCommandShowArgs {
			model, err := LoadModel(&file)

			if err != nil {
				fmt.Printf("Failed to load model %s: %v", file.Name(), err.Error())
				return
			}

			fmt.Printf("Model name: %s\n"+
				"Model word count: %d\n",
				model.Name, len(model.Words))

		}
		return
	}
	if modelCommandGenerate.Happened() {
		wordModel, err := LoadModel(modelCommandModelFileArg)

		if err != nil {
			fmt.Println("Failed to load model " + modelCommandModelFileArg.Name())
			return
		}
		fmt.Printf("Loaded %d words from model %s\n", len(wordModel.Words), wordModel.Name)

		fmt.Println(GenerateWords(wordModel, modelCommandGenerateCountArg))
		return
	}

	// handle config commands
	if configCommandShow.Happened() {
		if err := ConfigShowConfig(configCommandShowConfigFile); err != nil {
			fmt.Printf("Failed to show config %s: %v\n", configCommandShowConfigFile.Name(), err)
		}
		return
	}
	if configCommandCreate.Happened() {
		if err := ConfigCreateEmptyConfig(configCommandCreateConfigFile); err != nil {
			fmt.Printf("Failed to create a new config at %s: %v\n", configCommandCreateConfigFile.Name(), err)
		}
		return
	}
	if configCommandEdit.Happened() {
		ConfigEdit(configCommandEditConfigFile)
		return
	}

	// handle bot commands
	if botCommand.Happened() {
		if err := ConfigLoadConfig(botCommandRunConfigArg); err != nil {
			fmt.Printf("Failed to load config from %s: %s", botCommandRunConfigArg.Name(), err.Error())
			return
		}
		if err := RunBot(); err != nil {
			fmt.Printf("Error running bot: %v", err)
			return
		}
	}
}
