package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"math/rand"
	"os"
	"strconv"
)

func main() {
	parser := argparse.NewParser("Hurabotti", "Botin tynkä")

	// model commands
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
	modelCommandGenerateModelArg := modelCommandGenerate.File("m", "model", os.O_RDONLY, 0440, &argparse.Options{
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
	botCommandConfigFile := botCommandConfig.File("c", "config-file", os.O_RDWR|os.O_CREATE, 0660, botCommandConfigOptions)

	botCommandConfigShow := botCommandConfig.NewCommand("show", "Show config")
	botCommandConfigEdit := botCommandConfig.NewCommand("edit", "Edit config file")
	//botCommandConfigEditFile := botCommandConfigEdit.File("f", "config-file", os.O_RDWR, 0660, botCommandConfigOptions)

	botCommandRun := botCommand.NewCommand("run", "Run the bot")
	botCommandRunConfigArg := botCommandRun.File("c", "config-file", os.O_RDONLY, 0660, botCommandConfigOptions)

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Println("Error: " + err.Error())
		fmt.Println("Type -h for usage info")
		return
	}

	// handle model commands
	if modelCommandCreate.Happened() {
		CreateModelInit(modelCommandCreateArgs)
	} else if modelCommandList.Happened() {

	} else if modelCommandRemove.Happened() {

	} else if modelCommandGenerate.Happened() {
		wordModel := LoadModel(modelCommandGenerateModelArg)
		fmt.Println("Loaded " + strconv.Itoa(len(wordModel.Words)) + " words  from model " + wordModel.Name)
		// shuffle the first word for more randomness
		randomPosition := rand.Intn(len(wordModel.Words))
		firstWord := wordModel.Words[0]
		randomWord := wordModel.Words[randomPosition]

		wordModel.Words[0] = randomWord
		wordModel.Words[randomPosition] = firstWord

		generatedText := GenerateWords(wordModel, modelCommandGenerateCountArg)
		fmt.Println(generatedText)
	}

	// handle bot commands
	if botCommandConfigShow.Happened() {
		BotShowConfig(botCommandConfigFile)
	} else if botCommandConfigEdit.Happened() {
		//EditConfig(LoadConfig(botCommandConfigFile))
	} else if botCommandRun.Happened() {
		botConfig := LoadConfig(botCommandRunConfigArg)
		if err := RunBot(botConfig); err != nil {
			fmt.Println(err)
		}
	}
}
