package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
)

func CreateModel(files *[]string) {
	var filesToProcess []string

	for _, file := range *files {
		fileInfo, err := os.Stat(file)

		if err != nil {
			log.Fatal("Failed to read file or directory " + file + ": " + err.Error())
		}

		if fileInfo.IsDir() == true {
			filesInDirectory, err := os.ReadDir(file)

			if err != nil {
				log.Fatal("Failed to list files in directory " + file + ": " + err.Error())
			}

			for _, fileInDirectory := range filesInDirectory {
				if strings.HasSuffix(fileInDirectory.Name(), ".csv") {
					log.Print("[DEBUG] Found .csv file " + fileInDirectory.Name() + "in folder " + file)
					filesToProcess = append(filesToProcess, file+fileInDirectory.Name())
				} else {
					log.Print("[DEBUG] File " + fileInDirectory.Name() + " in folder " + file + " doesn't end with .csv, not processing")
					continue
				}
			}
		} else {
			if strings.HasSuffix(file, ".csv") {
				log.Print("[DEBUG] Found .csv file " + fileInfo.Name())
				filesToProcess = append(filesToProcess, file)
			} else {
				log.Print("[DEBUG] File " + file + "doesn't end with .csv, not processing")
				continue
			}
		}
	}

	processCSV(filesToProcess)
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
