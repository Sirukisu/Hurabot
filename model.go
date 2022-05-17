package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/dixonwille/skywalker"
	"github.com/mb-14/gomarkov"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// DiscordMessagesChannelInfoFromFile data decoded from channel.json files
type DiscordMessagesChannelInfoFromFile struct {
	// ID of channel, used later for organizing
	ID string
	// Type of channel, used later for finding groups
	Type int
	// Name of channel, used later for organizing
	Name string
	// Guild associated with the channel
	Guild DiscordMessagesChannelGuildInfoFromFile
}

// DiscordMessagesChannelGuildInfoFromFile guild data decoded from channel.json files
type DiscordMessagesChannelGuildInfoFromFile struct {
	// ID of guild, used later for organizing
	ID string
	// Name of guild, used later for organizing
	Name string
}

// DiscordGuild a Discord guild & it's channels
type DiscordGuild struct {
	// ID of guild
	ID int
	// Name of guild
	Name string
	// Slice of channels the guild contains
	Channels []DiscordChannel
}

// DiscordChannel a Discord channel
type DiscordChannel struct {
	// ID of channel
	ID int
	// Name of channel
	Name string
	// Is the channel enabled for model creation
	Enabled bool
}

// MessagesCsv data decoded from messages.csv files
type MessagesCsv struct {
	ID          int
	Timestamp   string
	Contents    string
	Attachments string
}

// WordModel containing a list of words
type WordModel struct {
	// Name of model
	Name string
	// Slice of words the model contains
	Words []string
}

// ChannelWorker Worker for reading channel directories in Discord message data
type ChannelWorker struct {
	*sync.Mutex
	found []string
}

// Work Function for skywalker for finding channel.json files in subdirectories
func (w *ChannelWorker) Work(path string) {
	w.Lock()
	defer w.Unlock()

	file, err := os.Stat(path)
	if err != nil {
		log.Println("Failed reading file " + path + ": " + err.Error())
	}

	if file.Name() == "channel.json" {
		w.found = append(w.found, path)
	}
}

// DiscordGuilds slice of loaded guilds
var DiscordGuilds = make([]DiscordGuild, 0)

// ModelFileName Filename of the model to be created
var ModelFileName string

// ModelName Name of the model to be created
var ModelName string

func CreateModel(directory *os.File) error {
	// try to load config from default location
	configLoaded := false

	if err := ConfigLoadConfig(nil); err == nil {
		configLoaded = true
	}

	// determine if we have a file or a directory

	fileInfo, err := directory.Stat()

	if err != nil {
		return fmt.Errorf("failed to get info from %s: %v", directory.Name(), err)
	}

	if fileInfo.IsDir() == false {
		return fmt.Errorf("%s is not a directory", directory.Name())
	}

	// check that directory has the index.json file
	_, err = os.Stat(path.Join(directory.Name(), "index.json"))

	if err != nil {
		return fmt.Errorf("failed to stat index.json file from %s: %v", directory.Name(), err)
	}

	// load the raw channel info
	channelInfo, err := LoadChannels(directory)
	if err != nil {
		return fmt.Errorf("failed to read channels from %s: %v", directory.Name(), err)
	}

	// open index.json file for decoding direct messages
	indexFile, err := os.Open(path.Join(directory.Name(), "index.json"))
	if err != nil {
		return fmt.Errorf("failed to open index file from %s: %v", directory.Name(), err)
	}

	dmInfo, err := LoadDirectMessages(indexFile)
	if err != nil {
		return fmt.Errorf("failed to read direct messages from %s: %v", directory.Name(), err)
	}

	// create the guilds
	DiscordGuilds, err = CreateGuilds(channelInfo, dmInfo)
	if err != nil {
		return fmt.Errorf("failed to create guilds from channel infos: %v", err)
	}

	if err := indexFile.Close(); err != nil {
		log.Printf("Failed to close index file %s: %v", indexFile.Name(), err)
	}

	// check that some guilds were loaded
	if len(DiscordGuilds) < 1 {
		return fmt.Errorf("no guilds found")
	}

	// start CUI for selecting enabled channels
	DiscordChannelSelectionCUI()

	// report model name and enabled channels after GUI
	fmt.Printf("Model filename: %s\n"+
		"Model name: %s\n"+
		"Enabled channels:\n",
		ModelFileName, ModelName)

	for _, guild := range DiscordGuilds {
		for _, channel := range guild.Channels {
			if channel.Enabled == true {
				fmt.Println(channel.Name)
			}
		}
	}

	// ask if user wants to start making the model
	fmt.Println("Continue? y/n")
	reader := bufio.NewReader(os.Stdin)
	resultChar, _, err := reader.ReadRune()

	if err != nil {
		log.Fatal(err)
	}

	if strings.ToLower(string(resultChar)) != "y" {
		return fmt.Errorf("aborted by user")
	}

	// check model name & modify is necessary
	if ModelName == "" {
		ModelName = "model"
		log.Println("Model name was not set, automatically setting it to " + ModelName)
	}

	if ModelFileName == "" {
		ModelFileName = "model"
		log.Println("Model filename was not set, automatically setting it to " + ModelFileName)
	}

	if strings.HasSuffix(ModelFileName, ".gob") == false {
		ModelFileName = ModelFileName + ".gob"
	}

	log.Println("Making model " + ModelName)

	var messagesParsed []MessagesCsv

	// parse the messages.csv files for all enabled channels
	for _, guild := range DiscordGuilds {
		for _, channel := range guild.Channels {
			if channel.Enabled == true {
				log.Printf("Processing channel %s in guild %s\n", channel.Name, guild.Name)

				// get the filepath of the channel's messages.csv
				messagesFilePath := path.Join(directory.Name(), fmt.Sprintf("c%d", channel.ID), "messages.csv")

				// open the messages.csv of the channel
				messagesCsv, err := os.Open(messagesFilePath)

				if err != nil {
					return fmt.Errorf("failed to open messages.csv file for channel %s: %v", channel.Name, err)
				}

				parsedMessages, err := ProcessMessagesCSV(messagesCsv)

				if err != nil {
					log.Printf("Failed to parse messages from channel %s: %v\n", channel.Name, err)
					continue
				}

				if err := messagesCsv.Close(); err != nil {
					log.Printf("Failed to close file %s: %v\n", messagesCsv.Name(), err)
				}

				messagesParsed = append(messagesParsed, parsedMessages...)

			}
		}
	}

	// close the directory file since it's no longer needed
	if err := directory.Close(); err != nil {
		log.Printf("Failed to close directory %s: %v\n", directory.Name(), err)
	}

	// check if any messages were parsed
	if len(messagesParsed) < 1 {
		log.Panicln("No messages were parsed, aborting")
	}

	log.Printf("Parsed %d total messages\n", len(messagesParsed))

	log.Println("Now sanitizing messages and splitting words")
	wordList := SanitizeMessages(messagesParsed)

	// check that wordList is not empty
	if len(wordList) < 1 {
		return fmt.Errorf("no messages were found")
	}

	// save model
	var saveDirectory string

	if configLoaded == true && LoadedConfig.ModelDirectory != "" {
		saveDirectory = path.Join(LoadedConfig.ModelDirectory)
	} else {
		wd, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to find the executable directory for model saving location: %v", err)
		}
		saveDirectory = path.Join(path.Dir(wd), "models")
	}

	log.Printf("Word processing done, now saving model to %s\n", path.Join(saveDirectory, ModelFileName))

	// check if models folder exists, create if not
	_, err = os.Stat(saveDirectory)

	if os.IsNotExist(err) {
		if err := os.MkdirAll(saveDirectory, 0770); err != nil {
			return fmt.Errorf("failed to create models directory at %s: %v", saveDirectory, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to open models directory at %s: %v", saveDirectory, err)
	}

	// check if model with same name already exists, create & open model file
	var fileExists bool
	var modelFile *os.File

	_, err = os.Stat(path.Join(saveDirectory, ModelFileName))

	if err == nil {
		fileExists = true
	} else if os.IsNotExist(err) {
		fileExists = false
	}

	if fileExists == false {
		// file doesn't exist, create
		modelFile, err = os.OpenFile(path.Join(saveDirectory, ModelFileName), os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			return fmt.Errorf("failed to open model file %s for writing: %v", ModelFileName, err)
		}
	} else {
		// file exists, ask user if it's ok to overwrite
		fmt.Println("Model " + ModelFileName + " already exists, overwrite? y/n")
		reader = bufio.NewReader(os.Stdin)
		resultChar, _, err := reader.ReadRune()

		if err != nil {
			return fmt.Errorf("invalid input, only use y/n")
		}

		switch strings.ToLower(string(resultChar)) {
		case "y":
			modelFile, err = os.OpenFile(path.Join(saveDirectory, ModelFileName), os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				return fmt.Errorf("failed to open model file %s for writing: %v", ModelFileName, err)
			}
		case "n":
			return fmt.Errorf("user aborted model creation")
		default:
			return fmt.Errorf("invalid input, only use y/n")
		}
	}

	// finally encode & save model to file
	enc := gob.NewEncoder(modelFile)
	if err := enc.Encode(WordModel{ModelName, wordList}); err != nil {
		return fmt.Errorf("failed to encode data to model file %s: %v", modelFile.Name(), err)
	}

	if err := modelFile.Close(); err != nil {
		log.Printf("Failed to close model file %s: %v\n", modelFile.Name(), err)
	}

	return nil
}

func LoadChannels(directory *os.File) ([]DiscordMessagesChannelInfoFromFile, error) {
	channelInfo := make([]DiscordMessagesChannelInfoFromFile, 0)

	// create the channel worker for Skywalker
	cw := new(ChannelWorker)
	cw.Mutex = new(sync.Mutex)

	// use Skywalker module to check every subdirectory for .json files
	sw := skywalker.New(directory.Name(), cw)
	sw.ExtListType = skywalker.LTWhitelist
	sw.ExtList = []string{".json"}
	sw.FilesOnly = true

	err := sw.Walk()
	if err != nil {
		return nil, fmt.Errorf("failed to read subdirectories: %v", err)
	}

	sort.Sort(sort.StringSlice(cw.found))

	for _, cf := range cw.found {
		file, err := os.Open(cf)
		if err != nil {
			log.Printf("Failed to open file %s: %v\n", cf, err)
			continue
		}

		newChannel := DiscordMessagesChannelInfoFromFile{}

		dec := json.NewDecoder(file)

		if err = dec.Decode(&newChannel); err != nil {
			log.Printf("Failed to decode file %s: %v\n", file.Name(), err)
		}

		if newChannel.Name != "" {
			channelInfo = append(channelInfo, newChannel)
		}

		if err := file.Close(); err != nil {
			log.Printf("Failed to close file %s: %v\n", file.Name(), err)
		}
	}

	return channelInfo, nil
}

func LoadDirectMessages(indexFile *os.File) (DiscordGuild, error) {
	directMessagesDecoded := map[string]string{}
	directMessagesGuild := DiscordGuild{
		ID:       1,
		Name:     "Direct Messages",
		Channels: nil,
	}

	dec := json.NewDecoder(indexFile)
	if err := dec.Decode(&directMessagesDecoded); err != nil {
		return directMessagesGuild, fmt.Errorf("failed to decode file %s: %v", indexFile.Name(), err)
	}

	for key, value := range directMessagesDecoded {
		if strings.HasPrefix(value, "Direct Message with") {
			channelID, err := strconv.Atoi(key)
			if err != nil {
				return directMessagesGuild, fmt.Errorf("failed to convert channel ID %s to integer: %v", key, err)
			}

			newDirectMessage := DiscordChannel{
				ID:      channelID,
				Name:    value,
				Enabled: false,
			}

			directMessagesGuild.Channels = append(directMessagesGuild.Channels, newDirectMessage)
		}
	}

	return directMessagesGuild, nil
}

func CreateGuilds(channelData []DiscordMessagesChannelInfoFromFile, directMessagesGuild DiscordGuild) ([]DiscordGuild, error) {
	// create list of guilds from channels
	discordGuilds := make([]DiscordGuild, 0)
	listOfGuilds := make([]DiscordMessagesChannelGuildInfoFromFile, 0)

	for _, channel := range channelData {
		newGuild := DiscordMessagesChannelGuildInfoFromFile{
			ID:   channel.Guild.ID,
			Name: channel.Guild.Name,
		}
		listOfGuilds = append(listOfGuilds, newGuild)
	}

	// remove duplicate guilds
	listOfGuildsParsed := CheckForDuplicateGuilds(listOfGuilds)

	// organize guilds with channels
	for _, guild := range listOfGuildsParsed {
		guildID, err := strconv.Atoi(guild.ID)
		if err != nil {
			guildID = 0
		}

		if guild.Name == "" {
			continue
		}

		newGuild := DiscordGuild{
			ID:       guildID,
			Name:     guild.Name,
			Channels: nil,
		}

		for _, channel := range channelData {
			if channel.Guild.ID == guild.ID {
				channelID, err := strconv.Atoi(channel.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to convert channel ID %s to integer: %v", channel.ID, err)
				}

				newChannel := DiscordChannel{
					ID:      channelID,
					Name:    channel.Name,
					Enabled: false,
				}

				newGuild.Channels = append(newGuild.Channels, newChannel)
			}
		}
		discordGuilds = append(discordGuilds, newGuild)
	}

	// add groups
	groupMessagesGuild := DiscordGuild{
		ID:       0,
		Name:     "Groups",
		Channels: nil,
	}

	// check for groups which have a type value of 3
	for _, channel := range channelData {
		if channel.Type == 3 {
			newGroupID, err := strconv.Atoi(channel.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to convert channel ID %s to integer: %v", channel.ID, err)
			}

			newGroup := DiscordChannel{
				ID:      newGroupID,
				Name:    channel.Name,
				Enabled: false,
			}

			groupMessagesGuild.Channels = append(groupMessagesGuild.Channels, newGroup)
		}
	}

	// check if any groups were found & append if yes
	if groupMessagesGuild.Channels != nil {
		discordGuilds = append(discordGuilds, groupMessagesGuild)
	}

	// same with DMs
	if directMessagesGuild.Channels != nil {
		discordGuilds = append(discordGuilds, directMessagesGuild)
	}

	return discordGuilds, nil
}

// CheckForDuplicateGuilds checks & removes duplicate guilds from a []DiscordMessagesChannelGuildInfoFromFile
func CheckForDuplicateGuilds(guilds []DiscordMessagesChannelGuildInfoFromFile) []DiscordMessagesChannelGuildInfoFromFile {
	occurred := map[string]bool{}
	result := make([]DiscordMessagesChannelGuildInfoFromFile, 0)

	for i := range guilds {
		if occurred[guilds[i].ID] != true {
			occurred[guilds[i].ID] = true
			result = append(result, guilds[i])
		}
	}
	return result
}

func ProcessMessagesCSV(csvFile *os.File) ([]MessagesCsv, error) {
	var parsedMessages []MessagesCsv

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = 4

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV file %s: %v", csvFile.Name(), err)
		}

		// skip the first line which has the record info
		if record[0] == "ID" {
			continue
		}

		newMessageId, err := strconv.Atoi(record[0])

		if err != nil {
			return nil, fmt.Errorf("failed to convert message ID %s to integer: %v", record[0], err)
		}

		newMessage := MessagesCsv{
			ID:          newMessageId,
			Timestamp:   record[1],
			Contents:    record[2],
			Attachments: record[3],
		}

		parsedMessages = append(parsedMessages, newMessage)
	}

	if len(parsedMessages) < 1 {
		return nil, fmt.Errorf("no messages were parsed from the file %s", csvFile.Name())
	}

	return parsedMessages, nil
}

func SanitizeMessages(messages []MessagesCsv) []string {
	// separate strings into words
	var wordList []string

	// loop through all messages, separate into words
	for _, message := range messages {
		messageWords := strings.Split(message.Contents, " ")

		for _, word := range messageWords {

			// word is empty, skip
			if word == "" {
				log.Println("Word is empty, skipping")
				continue
			}

			// check if word is a URL, skip if it is
			if strings.HasPrefix(word, "https://") || strings.HasPrefix(word, "http://") {
				log.Println("Word " + word + " is a URL, skipping")
				continue
			}

			// check if word has an animated emoji, then skip
			if strings.HasPrefix(word, "<a:") && strings.HasSuffix(word, ">") {
				log.Println("Word " + word + " is an animated emoji, skipping")
				continue
			}

			// word is a mention, skip
			if strings.Contains(word, "<@") && strings.HasSuffix(word, ">") {
				log.Println("Word " + word + " is a mention, skipping")
				continue
			}

			// word is a channel mention, skip
			if strings.HasPrefix(word, "<#") && strings.HasSuffix(word, ">") {
				log.Println("Word " + word + " is a channel mention, skipping")
				continue
			}

			// turn word to lowercase
			word = strings.ToLower(word)

			wordList = append(wordList, word)
		}
	}
	return wordList
}

// LoadModel loads a WordModel from os.File
func LoadModel(modelFile *os.File) (*WordModel, error) {
	var wordModel *WordModel

	dec := gob.NewDecoder(modelFile)
	err := dec.Decode(&wordModel)

	if err != nil {
		return nil, err
	}

	return wordModel, nil
}

// GenerateWords generates random words from a WordModel
func GenerateWords(model *WordModel, amount *int) string {
	// shuffle the first word for more randomness
	randomPosition := rand.Intn(len(model.Words))
	firstWord := model.Words[0]
	randomWord := model.Words[randomPosition]

	model.Words[0] = randomWord
	model.Words[randomPosition] = firstWord

	// create new chain
	chain := gomarkov.NewChain(1)

	// insert words to chain
	chain.Add(model.Words)

	tokens := make([]string, 0, *amount)
	tokens = append(tokens, gomarkov.StartToken)
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		if len(tokens) >= *amount+1 {
			break
		}
		next, _ := chain.Generate(tokens[(len(tokens) - 1):])
		tokens = append(tokens, next)
	}

	generatedText := strings.Join(tokens, " ")
	_, generatedText, _ = strings.Cut(generatedText, "$ ")
	return generatedText
}
