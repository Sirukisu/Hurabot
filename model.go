package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

type DiscordMessagesChannelInfo struct {
	Id    string
	Type  int
	Name  string
	Guild DiscordMessagesChannelGuildInfo
}

type DiscordMessagesChannelGuildInfo struct {
	Id   string
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
	fmt.Println(loadedChannels)

	// make list of guilds
	listOfGuilds := make([]DiscordMessagesChannelGuildInfo, 0, len(loadedChannels))

	for _, channel := range loadedChannels {
		newGuild := DiscordMessagesChannelGuildInfo{
			Id:   channel.Guild.Id,
			Name: channel.Guild.Name,
		}

		listOfGuilds = append(listOfGuilds, newGuild)
	}
	fmt.Println(listOfGuilds)
}

func loadChannelInfoFromMessages(directory *os.File, directoryContents []os.DirEntry) []DiscordMessagesChannelInfo {
	loadedChannels := make([]DiscordMessagesChannelInfo, 0, len(directoryContents))

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

				newChannel := DiscordMessagesChannelInfo{}

				err = json.Unmarshal(fileContent, &newChannel)

				if err != nil {
					fmt.Println(err)
				}

				loadedChannels = append(loadedChannels, newChannel)
			}
		}
	}
	fmt.Println(loadedChannels)
	return loadedChannels
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
