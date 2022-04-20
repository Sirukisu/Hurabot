package main

import (
	"fmt"
	"os"
)

func BotShowConfig(configFile *os.File) {
	// load the config
	botConfig := LoadConfig(configFile)

	// print info
	fmt.Println("Config file: " + configFile.Name())
	fmt.Println("Discord authentication token: " + botConfig.AuthenticationToken)
	fmt.Println("Discord bot command prefix: " + botConfig.CommandPrefix)
	fmt.Println("Discord guild ID: " + botConfig.GuildID)
	fmt.Println("Speech model to use: " + botConfig.ModelToUse)
	fmt.Println("Log directory: " + botConfig.LogDir)
	fmt.Println("Logging level: " + botConfig.LogLevel)
}
