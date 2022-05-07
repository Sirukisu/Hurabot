package main

import (
	"errors"
	"fmt"
	"log"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"os"
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
				botPrintLog("Received command request from "+i.Member.User.Username+"#"+i.Member.User.Discriminator+": "+
					i.ApplicationCommandData().Name, logger)
			} else if i.User != nil {
				botPrintLog("Received command request from "+i.User.Username+"#"+i.User.Discriminator+": "+
					i.ApplicationCommandData().Name, logger)
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
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: msg,
				},
			})

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
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				})
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
	logFile, err := openLog(LoadedConfig)
	if err != nil {
		return errors.New("error starting bot: failed to open log file " + err.Error())
	}
	logger = log.New(logFile, "", log.Flags())

	// read the model directory contents & load the models found
	// TODO individual file mode
	modelDirectoryContents, err := os.ReadDir(LoadedConfig.ModelDirectory)

	if err != nil {
		return errors.New("Failed to read model folder " + LoadedConfig.ModelDirectory + ": " + err.Error())
	}

	for _, file := range modelDirectoryContents {
		modelFile, err := os.Open(LoadedConfig.ModelDirectory + file.Name())
		if err != nil {
			return errors.New("Failed to open model " + file.Name() + ": " + err.Error())
		}
		wordModel := LoadModel(modelFile)
		botPrintLog("Loaded "+strconv.Itoa(len(wordModel.Words))+" words from model "+wordModel.Name, logger)
		wordModels = append(wordModels, wordModel)
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

	botPrintLog(strconv.Itoa(len(wordModels))+" models loaded in total", logger)

	// initialize the bot
	botPrintLog("Bot starting", logger)
	bot, err := discordgo.New("Bot " + LoadedConfig.AuthenticationToken)

	if err != nil {
		return errors.New("failed to create bot: " + err.Error())
	}

	botPrintLog("Adding handlers", logger)
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		botPrintLog("Logged in as: "+s.State.User.Username+"#"+s.State.User.Discriminator, logger)
	})

	err = bot.Open()
	if err != nil {
		botPrintLog("Cannot open the session: "+err.Error(), logger)
		return err
	}

	// register the commands from botCommands
	botPrintLog("Adding commands...", logger)
	registeredCommands := make([]*discordgo.ApplicationCommand, len(botCommands))
	for i, v := range botCommands {
		cmd, err := bot.ApplicationCommandCreate(bot.State.User.ID, LoadedConfig.GuildID, v)
		if err != nil {
			botPrintLog("Cannot create command "+v.Name+": "+err.Error(), logger)
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
	botPrintLog("Removing commands...", logger)
	// // We need to fetch the commands, since deleting requires the command ID.
	registeredCommands, err = bot.ApplicationCommands(bot.State.User.ID, LoadedConfig.GuildID)
	if err != nil {
		logger.Fatalln("Could not fetch registered commands: " + err.Error())
	}

	for _, v := range registeredCommands {
		err := bot.ApplicationCommandDelete(bot.State.User.ID, LoadedConfig.GuildID, v.ID)
		if err != nil {
			logger.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	logger.Println("Gracefully shutting down.")
	return nil
}

// openLog opens a log file in the directory of MainBotConfig for writing
func openLog(config *MainBotConfig) (*os.File, error) {
	_, err := os.Stat(config.LogDir)

	if os.IsNotExist(err) {
		err := os.Mkdir(config.LogDir, 0775)
		if err != nil {
			return nil, errors.New("failed to create log directory " + config.LogDir + ": " + err.Error())
		}
		fmt.Println("Log directory " + config.LogDir + " not, found, created it")
	} else if err != nil {
		return nil, errors.New("error reading log directory " + config.LogDir + ": " + err.Error())
	}

	timeNow := time.Now()
	logFilename := "log_" + strconv.Itoa(timeNow.Year()) + "_" + strconv.Itoa(int(timeNow.Month())) + "_" + strconv.Itoa(timeNow.Day()) + ".log"

	logFile, err := os.OpenFile(config.LogDir+logFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0664)

	if err != nil {
		return nil, errors.New("failed to open/create log file: " + err.Error())
	}

	return logFile, nil
}

// botPrintLog prints the input to stdout and the log file
func botPrintLog(input string, logger *log.Logger) {
	logger.Println(input)
	log.Println(input)
}
