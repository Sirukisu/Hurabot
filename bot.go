package main

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"strconv"
	"time"
)

var (
	// loaded models
	wordModels = make([]*WordModel, 0)
	// logger for writing to log file
	logger = &log.Logger{}

	// choices for text generation model option
	generateTextModelChoices = make([]*discordgo.ApplicationCommandOptionChoice, 0)

	// slice of bot commands
	botCommands = []*discordgo.ApplicationCommand{
		{
			Name:        "generate-text",
			Description: "Generate random text",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "model",
					Description: "Model to use for generating text",
					Choices:     generateTextModelChoices,
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "words",
					Description: "Amount of words to generate",
					MaxValue:    200,
					Required:    false,
				},
			},
		},
	}
	// map of command handlers
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		// handler for generate-text command
		"generate-text": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options

			// check if message was sent from a guild or from a DM and log accordingly
			// TODO add given options to the log message
			if i.Member != nil {
				logger.Printf("Received command request from %s#%s: %v\n", i.Member.User.Username, i.Member.User.Discriminator,
					i.ApplicationCommandData())
			} else if i.User != nil {
				logger.Printf("Received command request from %s#%s: %v\n", i.User.Username, i.User.Discriminator,
					i.ApplicationCommandData())
			}

			// make map of the options received
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			msg := "Generating text with "

			if option, ok := optionMap["words"]; ok {
				msg += strconv.Itoa(int(option.IntValue())) + " words"
			} else {
				msg += "50 words"
			}

			msg += " using model " + wordModels[optionMap["model"].IntValue()].Name

			// send response
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg,
				},
			}); err != nil {
				logger.Printf("Failed to send interaction response: %v\n", err)
			}

			// generate the text
			var generatedText string
			var amountOfWords = 50

			// set value for amount of words if it was supplied
			if option, ok := optionMap["words"]; ok {
				amountOfWords = int(option.IntValue())
			}

			generatedText = GenerateWords(wordModels[optionMap["model"].IntValue()], &amountOfWords)

			// edit the response with the generated text
			_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: msg + "\n\n" + generatedText,
			})
			if err != nil {
				if _, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				}); err != nil {
					logger.Printf("Failed to send followup message: %v\n", err)
				}
				return
			}
		},
	}
)

// RunBot runs the Discord bot
func RunBot() error {
	// check that authentication token is set
	if LoadedConfig.AuthenticationToken == "" {
		return errors.New("error starting bot: authentication token is empty")
	}

	// initialize the logger
	logFile, err := openLog()
	if err != nil {
		return errors.New("error starting bot: failed to open log file " + err.Error())
	}
	logMultiWriter := io.MultiWriter(logFile, os.Stdout)
	logger = log.New(logMultiWriter, "", log.Flags())

	// read the model directory contents & load the models found
	// check if models are set in config
	if len(LoadedConfig.ModelsToUse) > 0 {
		for i := range LoadedConfig.ModelsToUse {
			var modelFile *os.File

			modelFile, err = os.Open(path.Clean(LoadedConfig.ModelsToUse[i]))

			if err != nil {
				// try to find in config models dir
				modelFile, err = os.Open(path.Join(LoadedConfig.ModelDirectory, LoadedConfig.ModelsToUse[i]))

				if err != nil {
					logger.Printf("Failed to load model file %s: %v\n", LoadedConfig.ModelsToUse[i], err)
					continue
				}
			}

			wordModel, err := LoadModel(modelFile)
			if err != nil {
				logger.Printf("Failed to load model from file %s: %v\n", modelFile.Name(), err)
			}

			logger.Printf("Loaded %d words from model %s\n", len(wordModel.Words), wordModel.Name)

			if err := modelFile.Close(); err != nil {
				logger.Printf("Failed to close model file %s: %v", modelFile.Name(), err)
			}

			wordModels = append(wordModels, wordModel)
		}
	} else {
		// Load whole directory
		modelDirectoryContents, err := os.ReadDir(LoadedConfig.ModelDirectory)

		if err != nil {
			return errors.New("Failed to read model folder " + LoadedConfig.ModelDirectory + ": " + err.Error())
		}

		for _, file := range modelDirectoryContents {
			modelFile, err := os.Open(LoadedConfig.ModelDirectory + file.Name())
			if err != nil {
				return errors.New("failed to open model " + file.Name() + ": " + err.Error())
			}
			wordModel, err := LoadModel(modelFile)
			if err != nil {
				return errors.New("failed to decode model " + modelFile.Name() + ": " + err.Error())
			}
			logger.Printf("Loaded %d words from model %s\n", len(wordModel.Words), wordModel.Name)
			wordModels = append(wordModels, wordModel)
		}
	}

	// check that some models were loaded
	if len(wordModels) < 1 {
		return errors.New("no word models were loaded")
	}

	// generate the model option choices from loaded models & update the variable
	for i, model := range wordModels {
		textGenerationModelChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  model.Name,
			Value: i,
		}
		generateTextModelChoices = append(generateTextModelChoices, textGenerationModelChoice)
	}

	// we know that text-generate command and the model option are both index 0
	botCommands[0].Options[0].Choices = generateTextModelChoices

	// also set the max amount of words from config
	botCommands[0].Options[1].MaxValue = float64(LoadedConfig.MaxWords)

	logger.Printf("%d models loaded in total\n", len(wordModels))

	// initialize the bot
	logger.Println("Bot starting")
	bot, err := discordgo.New("Bot " + LoadedConfig.AuthenticationToken)

	if err != nil {
		return errors.New("failed to create bot: " + err.Error())
	}

	logger.Println("Adding handlers")
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		logger.Printf("Logged in as %s#%s\n", s.State.User.Username, s.State.User.Discriminator)
	})

	if err := bot.Open(); err != nil {
		return fmt.Errorf("cannot open the session: %v", err)
	}

	// register the commands from botCommands
	logger.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(botCommands))
	for i, v := range botCommands {
		cmd, err := bot.ApplicationCommandCreate(bot.State.User.ID, LoadedConfig.GuildID, v)
		if err != nil {
			logger.Printf("Cannot create command %s: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer bot.Close()

	// shutdown bot after Ctrl+C is received
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	// remove the commands & shut down
	logger.Println("Removing commands...")
	registeredCommands, err = bot.ApplicationCommands(bot.State.User.ID, LoadedConfig.GuildID)
	if err != nil {
		logger.Fatalf("Could not fetch registered commands: %v", err)
	}

	for _, v := range registeredCommands {
		err := bot.ApplicationCommandDelete(bot.State.User.ID, LoadedConfig.GuildID, v.ID)
		if err != nil {
			logger.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	logger.Println("Gracefully shutting down.")

	logger = nil
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file %s: %v", logFile.Name(), err)
	}
	return nil
}

// openLog opens a log file in the directory of MainBotConfig for writing
func openLog() (*os.File, error) {
	if LoadedConfig.LogDir == "" {
		return nil, errors.New("no log dir found in config")
	}

	_, err := os.Stat(LoadedConfig.LogDir)

	if os.IsNotExist(err) {
		err := os.MkdirAll(LoadedConfig.LogDir, 0775)
		if err != nil {
			return nil, errors.New("failed to create log directory " + LoadedConfig.LogDir + ": " + err.Error())
		}
		fmt.Printf("Log directory %s not, found, created it\n", LoadedConfig.LogDir)
	} else if err != nil {
		return nil, errors.New("error reading log directory " + LoadedConfig.LogDir + ": " + err.Error())
	}

	timeNow := time.Now()
	logFilename := fmt.Sprintf("log_%d_%d_%d.log", timeNow.Year(), timeNow.Month(), timeNow.Day())

	logFile, err := os.OpenFile(LoadedConfig.LogDir+logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)

	if err != nil {
		return nil, errors.New("failed to open/create log file: " + err.Error())
	}

	return logFile, nil
}
