package main

import (
	"fmt"
	"github.com/akamensky/argparse"
	"log"
	"os"
	"strings"
)

func main() {
	parser := argparse.NewParser("Hurabotti", "Botin tynkä")
	modelCommand := parser.NewCommand("model", "manage bot word models")
	modelCommandCreate := modelCommand.NewCommand("create", "create new model")
	modelCommandCreate.StringList("f", "file", &argparse.Options{
		Required: true,
		Validate: verifyCsv,
		Help:     "CSV file or folder to process",
		Default:  nil,
	})

	modelCommandList := modelCommand.NewCommand("list", "List current models")
	modelCommandRemove := modelCommand.NewCommand("remove", "Remove a model")

	err := parser.Parse(os.Args)

	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	if modelCommandCreate.Happened() {
		fmt.Println("Stuff for model creating goes here")
	} else if modelCommandList.Happened() {

	} else if modelCommandRemove.Happened() {

	}
}

func verifyCsv(args []string) error {
	fileInfo, err := os.Stat(args[0])

	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir() == false {
		if strings.HasSuffix(args[0], ".csv") == false {
			log.Fatal("File doesn't end with .csv.")
		}
	}
	return nil
}
