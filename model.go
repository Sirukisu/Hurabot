package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dixonwille/skywalker"
	"github.com/mb-14/gomarkov"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type DiscordMessagesChannelInfoFromFile struct {
	Id    string
	Type  int
	Name  string
	Guild DiscordMessagesChannelGuildInfoFromFile
}

type DiscordMessagesChannelGuildInfoFromFile struct {
	Id   string
	Name string
}

type DiscordGuild struct {
	Id       int
	Name     string
	Channels []DiscordChannel
}

type DiscordChannel struct {
	Id      int
	Name    string
	Enabled bool
}

type MessagesCsv struct {
	ID          int
	Timestamp   string
	Contents    string
	Attachments string
}

type WordModel struct {
	Words []string
}

type ChannelWorker struct {
	*sync.Mutex
	found []string
}

func (ew *ChannelWorker) Work(path string) {
	ew.Lock()
	defer ew.Unlock()

	file, err := os.Stat(path)
	if err != nil {
		log.Println("Failed reading file " + path + ": " + err.Error())
	}

	if file.Name() == "channel.json" {
		ew.found = append(ew.found, path)
	}
}

var DiscordGuilds = make([]DiscordGuild, 0)
var ModelName string

func CreateModelInit(directory *os.File) {
	directoryInfo, err := directory.Stat()

	if err != nil {
		fmt.Println(err)
		return
	}

	if directoryInfo.IsDir() != true {
		fmt.Println("File " + directory.Name() + " is not a directory")
		return
	}

	// just load everything from the directory & subdirectories
	loadedChannels := loadChannelInfoFromMessages(directory)

	// make list of guilds & check for duplicates
	listOfGuilds := make([]DiscordMessagesChannelGuildInfoFromFile, 0)

	for _, channel := range loadedChannels {
		newGuild := DiscordMessagesChannelGuildInfoFromFile{
			Id:   channel.Guild.Id,
			Name: channel.Guild.Name,
		}

		listOfGuilds = append(listOfGuilds, newGuild)
	}
	listOfGuildsParsed := checkForDuplicateGuilds(listOfGuilds)

	// finally organize guilds with channels

	for _, guild := range listOfGuildsParsed {
		guildId, err := strconv.Atoi(guild.Id)

		if err != nil {
			guildId = 0
		}

		if guild.Name == "" {
			continue
		}

		newGuild := DiscordGuild{
			Id:       guildId,
			Name:     guild.Name,
			Channels: nil,
		}

		for _, channel := range loadedChannels {
			if channel.Guild.Id == guild.Id {
				channelID, err := strconv.Atoi(channel.Id)

				if err != nil {
					fmt.Println(err)
				}

				newChannel := DiscordChannel{
					Id:      channelID,
					Name:    channel.Name,
					Enabled: false,
				}

				newGuild.Channels = append(newGuild.Channels, newChannel)
			}
		}
		DiscordGuilds = append(DiscordGuilds, newGuild)
	}

	// fix direct messages & groups
	groupMessagesGuild := DiscordGuild{
		Id:       1,
		Name:     "Groups",
		Channels: nil,
	}

	// check for groups which have a type value of 3
	for _, channel := range loadedChannels {
		if channel.Type == 3 {
			newGroupId, err := strconv.Atoi(channel.Id)

			if err != nil {
				log.Println("Failed to parse channel " + channel.Name + " ID: " + err.Error())
			}

			newGroup := DiscordChannel{
				Id:      newGroupId,
				Name:    channel.Name,
				Enabled: false,
			}

			groupMessagesGuild.Channels = append(groupMessagesGuild.Channels, newGroup)
		}
	}

	directMessagesGuild := DiscordGuild{
		Id:       0,
		Name:     "Direct Messages",
		Channels: nil,
	}

	// find the direct message names from index.json file
	index, err := os.ReadFile(directory.Name() + string(os.PathSeparator) + "index.json")

	if err != nil {
		fmt.Println(err)
	}

	marshalledIndex := map[string]string{}
	err = json.Unmarshal(index, &marshalledIndex)

	if err != nil {
		fmt.Println(err)
	}

	for key, value := range marshalledIndex {
		if strings.HasPrefix(value, "Direct Message with") {
			channelID, err := strconv.Atoi(key)

			if err != nil {
				log.Fatal(err)
			}

			newDirectMessage := DiscordChannel{
				Id:      channelID,
				Name:    value,
				Enabled: false,
			}
			directMessagesGuild.Channels = append(directMessagesGuild.Channels, newDirectMessage)
		}
	}

	// check if the guilds have channels, append if they do
	if groupMessagesGuild.Channels != nil {
		DiscordGuilds = append(DiscordGuilds, groupMessagesGuild)
	}
	if directMessagesGuild.Channels != nil {
		DiscordGuilds = append(DiscordGuilds, directMessagesGuild)
	}

	// start GUI for selecting enabled channels
	DiscordChannelSelectionGUI()

	// report model name and enabled channels after GUI
	fmt.Println("Model name: " + ModelName)
	fmt.Println("Enabled channels:")

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

	switch strings.ToLower(string(resultChar)) {
	case "y":
		CreateModel(directory)
	case "n":
		fmt.Println("Aborted")
	}
}

func loadChannelInfoFromMessages(directory *os.File) []DiscordMessagesChannelInfoFromFile {
	channelInfo := make([]DiscordMessagesChannelInfoFromFile, 0)
	ew := new(ChannelWorker)
	ew.Mutex = new(sync.Mutex)

	sw := skywalker.New(directory.Name(), ew)
	sw.ExtListType = skywalker.LTWhitelist
	sw.ExtList = []string{".json"}
	sw.FilesOnly = true

	err := sw.Walk()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	sort.Sort(sort.StringSlice(ew.found))
	for _, f := range ew.found {
		channelData, err := os.ReadFile(f)

		if err != nil {
			log.Println("Failed to read file " + f + ": " + err.Error())
			continue
		}
		newChannel := DiscordMessagesChannelInfoFromFile{}

		if err = json.Unmarshal(channelData, &newChannel); err != nil {
			log.Println("Failed to decode file " + f + ": " + err.Error())
		}

		if newChannel.Name != "" {
			channelInfo = append(channelInfo, newChannel)
		}
	}

	return channelInfo
}

func checkForDuplicateGuilds(guilds []DiscordMessagesChannelGuildInfoFromFile) []DiscordMessagesChannelGuildInfoFromFile {
	occurred := map[string]bool{}
	result := make([]DiscordMessagesChannelGuildInfoFromFile, 0)

	for i := range guilds {
		if occurred[guilds[i].Id] != true {
			occurred[guilds[i].Id] = true
			result = append(result, guilds[i])
		}
	}
	return result
}

func CreateModel(fileOrDirectory *os.File) {
	// check that there are some guilds
	if len(DiscordGuilds) < 1 {
		log.Fatalln("No guilds loaded, aborting")
	}

	// check model name & modify is necessary
	if ModelName == "" {
		ModelName = "model"
		log.Println("Model name was not set, automatically setting it to " + ModelName)
	}

	if strings.HasSuffix(ModelName, ".gob") == false {
		ModelName = ModelName + ".gob"
	}

	log.Println("Making model " + ModelName)

	// create parsed messages variable
	var messagesParsed []MessagesCsv

	// find if we have a file or a directory
	fileInfo, err := fileOrDirectory.Stat()

	if err != nil {
		log.Fatalln("Could not read " + fileOrDirectory.Name() + ": " + err.Error())
	}

	if fileInfo.IsDir() == true {
		// directory processing mode

		if err != nil {
			log.Fatalln("Failed to read directory contents: " + err.Error())
		}

		// loop through all enabled channels & process their messages.csv files
		for _, guild := range DiscordGuilds {
			for _, channel := range guild.Channels {
				if channel.Enabled == true {
					log.Println("Processing channel " + channel.Name + " in guild " + guild.Name)

					// change into channel directory
					if err := os.Chdir(fileOrDirectory.Name() + string(os.PathSeparator) + "c" + strconv.Itoa(channel.Id)); err != nil {
						log.Fatalln("Failed to change to channel directory: " + err.Error())
					}

					// process messages.csv
					messagesCsv, err := os.Open("messages.csv")

					if err != nil {
						log.Fatalln("Failed to read messages.csv for channel " + channel.Name + ": " + err.Error())
					}

					// parse messages & append to the slice
					newMessagesParsed, err := processMessagesCSV(messagesCsv)

					if err != nil {
						log.Fatalln("Error parsing messages.csv for channel " + channel.Name + ": " + err.Error())
					}
					messagesParsed = append(messagesParsed, newMessagesParsed...)
				}
			}
		}

	} else {
		// TODO file processing mode
	}

	// check if any messages were parsed
	if len(messagesParsed) < 1 {
		log.Panicln("No messages were parsed, aborting")
	}
	log.Println("Parsed " + strconv.Itoa(len(messagesParsed)) + " total messages")

	// filter messages
	log.Println("Now sanitizing messages and splitting words")
	wordList := sanitizeMessages(messagesParsed)

	// save model
	log.Println("Word processing done, saving model to models/" + ModelName)

	err = saveModel(wordList)

	if err != nil {
		log.Fatalln(err)
	}
}

func processMessagesCSV(csvFile *os.File) ([]MessagesCsv, error) {
	if csvFile.Name() != "messages.csv" {
		return nil, errors.New("filename is not messages.csv")
	}

	var messages []MessagesCsv

	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = 4

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// skip the first line which has the record info
		if record[0] == "ID" {
			continue
		}

		newMessageId, err := strconv.Atoi(record[0])

		if err != nil {
			return nil, err
		}

		newMessage := MessagesCsv{
			ID:          newMessageId,
			Timestamp:   record[1],
			Contents:    record[2],
			Attachments: record[3],
		}

		messages = append(messages, newMessage)
	}
	return messages, nil
}

func sanitizeMessages(messages []MessagesCsv) []string {
	// separate strings into words & emojis
	wordList := make([]string, 0)

	// loop through all messages, separate into words & emojis
	for _, message := range messages {
		messageWords := strings.Split(message.Contents, " ")
		for _, word := range messageWords {

			// check if word is a URL, skip if it is
			_, err := url.ParseRequestURI(word)

			if err == nil {
				log.Println("Word " + word + " is a URL, skipping")
				continue
			}

			// check if word has an animated emoji, then skip
			if strings.HasPrefix(word, "<a:") && strings.HasSuffix(word, ">") {
				log.Println("Word " + word + " is an animated emoji, skipping")
				continue
			}

			// word is empty, skip
			if word == "" {
				log.Println("Word is empty, skipping")
				continue
			}

			// word is a mention, skip
			if strings.Contains(word, "<@!") && strings.HasSuffix(word, ">") {
				log.Println("Word " + word + " is a mention, skipping")
				continue
			}

			// turn word to lowercase
			word = strings.ToLower(word)

			wordList = append(wordList, word)
		}
	}
	return wordList
}

func saveModel(words []string) error {
	// check that words contain something
	if len(words) < 1 {
		return errors.New("word list is empty")
	}

	// open model file for writing
	programPath, err := os.Executable()

	if err != nil {
		return errors.New("Unable to find executable path: " + err.Error())
	}

	programPath, _ = filepath.Split(programPath)

	// check if models folder exists, create if not
	_, err = os.Stat(programPath + string(os.PathSeparator) + "models")

	if os.IsNotExist(err) {
		err := os.Mkdir(programPath+string(os.PathSeparator)+"models", 1775)
		if err != nil {
			return errors.New("Failed to create models directory at " + programPath + ": " + err.Error())
		}
	} else if err != nil {
		return errors.New("Failed to read models folder at " + programPath + ": " + err.Error())
	}

	// check if model with same name already exists, create & open model file, finally write model to file

	_, err = os.Stat(programPath + string(os.PathSeparator) + "models" + string(os.PathSeparator) + ModelName)

	if os.IsNotExist(err) {
		modelFile, err := os.OpenFile(programPath+string(os.PathSeparator)+"models"+string(os.PathSeparator)+ModelName, os.O_WRONLY|os.O_CREATE, 0664)
		if err != nil {
			return errors.New("Failed creating model file " + ModelName + ": " + err.Error())
		}
		enc := gob.NewEncoder(modelFile)
		err = enc.Encode(WordModel{words})

		if err != nil {
			return errors.New("Error encoding data: " + err.Error())
		}
	} else if err != nil {
		return errors.New("Failed to read file " + ModelName + ": " + err.Error())
	} else {
		// file already exists, ask if user wants to overwrite it
		fmt.Println("Model " + ModelName + " already exists, overwrite? y/n")
		reader := bufio.NewReader(os.Stdin)
		resultChar, _, err := reader.ReadRune()

		if err != nil {
			return err
		}

		switch strings.ToLower(string(resultChar)) {
		case "y":
			modelFile, err := os.OpenFile(programPath+string(os.PathSeparator)+"models"+string(os.PathSeparator)+ModelName, os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				return errors.New("Failed to create model file " + ModelName + ": " + err.Error())
			}

			enc := gob.NewEncoder(modelFile)
			err = enc.Encode(WordModel{words})

			if err != nil {
				return errors.New("Error encoding data: " + err.Error())
			}
		case "n":
			return errors.New("File " + ModelName + " already exists")
		default:
			return errors.New("invalid answer, only use y/n")
		}
	}

	return nil
}

func loadModel(modelFile *os.File) *WordModel {
	var wordModel *WordModel

	dec := gob.NewDecoder(modelFile)
	err := dec.Decode(&wordModel)

	if err != nil {
		log.Fatalln("Failed to decode model " + modelFile.Name() + ": " + err.Error())
	}

	return wordModel
}

func generateWords(model *WordModel, amount *int) {
	// create new chain
	chain := gomarkov.NewChain(1)

	// insert words to chain
	chain.Add(model.Words)

	tokens := make([]string, 0, *amount)
	tokens = append(tokens, gomarkov.StartToken)
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		if len(tokens) >= *amount {
			break
		}
		next, _ := chain.Generate(tokens[(len(tokens) - 1):])
		tokens = append(tokens, next)
	}

	generatedText := strings.Join(tokens, " ")
	_, generatedText, _ = strings.Cut(generatedText, "$ ")
	fmt.Println(generatedText)
}
