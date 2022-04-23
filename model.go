package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	Id   int
	Name string
}

func CreateModelFromMessages(directory *os.File) {
	directoryInfo, err := directory.Stat()

	if err != nil {
		fmt.Println(err)
		return
	}

	if directoryInfo.IsDir() != true {
		fmt.Println("File " + directory.Name() + " is not a directory")
		return
	}

	directoryContents, err := directory.ReadDir(0)

	if err != nil {
		fmt.Println("Failed to read contents of " + directory.Name() + ": " + err.Error())
		return
	}

	loadedChannels := loadChannelInfoFromMessages(directory, directoryContents)

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
	discordGuilds := make([]DiscordGuild, 0)

	for _, guild := range listOfGuildsParsed {
		guildId, err := strconv.Atoi(guild.Id)

		if err != nil {
			guildId = 0
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
					Id:   channelID,
					Name: channel.Name,
				}

				newGuild.Channels = append(newGuild.Channels, newChannel)
			}
		}
		discordGuilds = append(discordGuilds, newGuild)
	}

	// fix direct messages
	directMessagesGuild := DiscordGuild{
		Id:       0,
		Name:     "Direct Messages",
		Channels: nil,
	}

	index, err := os.ReadFile(directory.Name() + string(os.PathSeparator) + "index.json")

	if err != nil {
		fmt.Println(err)
	}

	marshalledIndex := map[string]string{}
	json.Unmarshal(index, &marshalledIndex)

	for key, value := range marshalledIndex {
		if strings.HasPrefix(value, "Direct Message with") {
			channelID, err := strconv.Atoi(key)

			if err != nil {
				log.Fatal(err)
			}

			newDirectMessage := DiscordChannel{
				Id:   channelID,
				Name: value,
			}
			directMessagesGuild.Channels = append(directMessagesGuild.Channels, newDirectMessage)
		}
	}

	discordGuilds = append(discordGuilds, directMessagesGuild)
}

func loadChannelInfoFromMessages(directory *os.File, directoryContents []os.DirEntry) []DiscordMessagesChannelInfoFromFile {
	loadedChannels := make([]DiscordMessagesChannelInfoFromFile, 0)

	for _, dc := range directoryContents {
		os.Chdir(directory.Name() + string(os.PathSeparator) + dc.Name())
		wd, err := os.Getwd()

		if err != nil {
			fmt.Println(err)
			continue
		}

		subDirectoryContents, _ := os.ReadDir(wd)
		for _, sdc := range subDirectoryContents {
			file, err := sdc.Info()

			if err != nil {
				fmt.Println(err)
			}

			if file.Name() == "channel.json" {
				fileContent, err := os.ReadFile(file.Name())

				if err != nil {
					fmt.Println(err)
				}

				newChannel := DiscordMessagesChannelInfoFromFile{}

				err = json.Unmarshal(fileContent, &newChannel)

				if err != nil {
					fmt.Println(err)
				}

				loadedChannels = append(loadedChannels, newChannel)
			}
		}
	}
	return loadedChannels
}

func checkForDuplicateGuilds(guilds []DiscordMessagesChannelGuildInfoFromFile) []DiscordMessagesChannelGuildInfoFromFile {
	occurred := map[string]bool{}
	result := make([]DiscordMessagesChannelGuildInfoFromFile, 0)

	for i, _ := range guilds {
		if occurred[guilds[i].Id] != true {
			occurred[guilds[i].Id] = true
			result = append(result, guilds[i])
		}
	}
	return result
}

func processCSV(files []string) {
	//var wordList []string

	for _, file := range files {
		fileContent, err := os.Open(file)

		if err != nil {
			fmt.Println("Error reading file " + file + ": " + err.Error())
		}

		reader := csv.NewReader(fileContent)
		records, _ := reader.ReadAll()
		fmt.Println(records)
	}
}
