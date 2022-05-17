package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestLoadChannels(t *testing.T) {
	testDir, err := os.MkdirTemp(os.TempDir(), "hurabotTestLoadChannels")

	if err != nil {
		t.Fatal(err)
	}

	testDirFile, err := os.Open(testDir)

	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		channelDir := path.Join(testDir, fmt.Sprintf("c%d", i))
		if err := os.Mkdir(channelDir, 0770); err != nil {
			t.Fatal(err)
		}
		channelJson := DiscordMessagesChannelInfoFromFile{
			ID:   strconv.Itoa(i),
			Type: 0,
			Name: fmt.Sprintf("Test channel %d", i),
			Guild: DiscordMessagesChannelGuildInfoFromFile{
				ID:   strconv.Itoa(i),
				Name: fmt.Sprintf("Test guild %d", i),
			},
		}

		channelJsonFile, err := os.Create(path.Join(channelDir, "channel.json"))

		if err != nil {
			t.Fatal(err)
		}

		enc := json.NewEncoder(channelJsonFile)
		if err := enc.Encode(channelJson); err != nil {
			t.Fatal(err)
		}

		if err := channelJsonFile.Close(); err != nil {
			t.Fatal(err)
		}
	}

	testChannels, err := LoadChannels(testDirFile)

	if err != nil {
		t.Errorf("error loading channels: %v", err)
	}

	if len(testChannels) != 3 && t.Failed() == false {
		t.Errorf("length of loaded channels was %d instead of 3", len(testChannels))
	}

	if err := testDirFile.Close(); err != nil {
		t.Logf("failed to close test directory %s: %v", testDir, err)
	}

	if err := os.RemoveAll(testDir); err != nil {
		t.Logf("failed to remove the test directory %s: %v", testDir, err)
	}
}

func TestLoadDirectMessages(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	testDir, err := os.MkdirTemp(os.TempDir(), "hurabotTestLoadDirectMessages")

	if err != nil {
		t.Fatal(err)
	}

	indexJsonFile, err := os.Create(path.Join(testDir, "index.json"))

	if err != nil {
		t.Fatal(err)
	}

	testDirectMessages := map[string]string{}

	for i := 0; i < 6; i++ {
		key := strconv.Itoa(rand.Int())
		var value string
		if i == 3 || i == 4 {
			value = "Direct Message with someone"
		} else {
			value = "Random guild name"
		}
		testDirectMessages[key] = value
	}

	enc := json.NewEncoder(indexJsonFile)
	enc.SetIndent(" ", "")
	if err := enc.Encode(testDirectMessages); err != nil {
		t.Fatal(err)
	}

	if err := indexJsonFile.Close(); err != nil {
		t.Logf("failed to close test file %s: %v", indexJsonFile.Name(), err)
	}

	indexJsonFile, err = os.Open(path.Join(testDir, "index.json"))

	if err != nil {
		t.Fatal(err)
	}

	loadedMessages, err := LoadDirectMessages(indexJsonFile)

	if err != nil {
		t.Errorf("failed loading direct messages from %s: %v", indexJsonFile.Name(), err)
	}

	if len(loadedMessages.Channels) != 2 && t.Failed() == false {
		t.Errorf("length of loaded direct messages was %d instead of 2", len(loadedMessages.Channels))
	}

	if err := indexJsonFile.Close(); err != nil {
		t.Logf("failed to close test file %s: %v", indexJsonFile.Name(), err)
	}

	if err := os.RemoveAll(testDir); err != nil {
		t.Logf("failed to remove the test directory %s: %v", testDir, err)
	}
}

func TestCreateGuilds(t *testing.T) {
	testData := make([]DiscordMessagesChannelInfoFromFile, 0)

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 9; i++ {
		newTestData := DiscordMessagesChannelInfoFromFile{
			ID:   strconv.Itoa(rand.Int()),
			Type: 0,
			Name: fmt.Sprintf("Test channel #%d", i),
			Guild: DiscordMessagesChannelGuildInfoFromFile{
				ID:   "123",
				Name: "Test Guild",
			},
		}
		testData = append(testData, newTestData)
	}

	directMessageChannel := DiscordChannel{
		ID:      0,
		Name:    "Direct Message with someone",
		Enabled: false,
	}

	directMessageGuild := DiscordGuild{
		ID:       0,
		Name:     "",
		Channels: []DiscordChannel{},
	}

	directMessageGuild.Channels = append(directMessageGuild.Channels, directMessageChannel)

	discordGuilds, err := CreateGuilds(testData, directMessageGuild)

	if err != nil {
		t.Fatalf("error creating guilds: %v", err)
	}

	if len(discordGuilds) != 2 && t.Failed() == false {
		t.Errorf("error creating guilds: guild length was %d instead of 2", len(discordGuilds))
	}
}

func TestProcessMessagesCSV(t *testing.T) {
	testFile, err := os.CreateTemp(os.TempDir(), "hurabotTestProcessMessagesCSV*.csv")

	if err != nil {
		t.Fatal(err)
	}

	testCsvData := `ID,Timestamp,Contents,Attachments
000000000000000001,2000-01-01 12:00:00.000000+00:00,Test message #1,
000000000000000002,2000-01-01 12:00:00.000000+00:00,Test message #2,
000000000000000003,2000-01-01 12:00:00.000000+00:00,Attachment message,https://test.attachment.test/
000000000000000004,2000-01-01 12:00:00.000000+00:00,https://test.attachment.test/,
000000000000000005,2000-01-01 12:00:00.000000+00:00,Wow animated emoji <a:123456789>,
000000000000000006,2000-01-01 12:00:00.000000+00:00,Hey <@123456789> this is a mention!,
000000000000000007,2000-01-01 12:00:00.000000+00:00,<#123456789> is a pretty cool channel,
`

	if _, err := testFile.WriteString(testCsvData); err != nil {
		t.Fatal(err)
	}

	if err := testFile.Close(); err != nil {
		t.Errorf("failed to close test file: %v", err)
	}

	testFile, err = os.Open(testFile.Name())

	if err != nil {
		t.Fatal(err)
	}

	parsedMessages, err := ProcessMessagesCSV(testFile)

	if err != nil {
		t.Errorf("failed to parse messages.csv: %v", err)
	}

	if len(parsedMessages) != 7 && t.Failed() == false {
		t.Errorf("error parsing messages.csv: parsed messages length was %d instead of 7", len(parsedMessages))
	}

	if err := testFile.Close(); err != nil {
		t.Logf("failed to close test file %s: %v", testFile.Name(), err)
	}

	if err := os.Remove(testFile.Name()); err != nil {
		t.Logf("failed to remove test file %s: %v", testFile.Name(), err)
	}
}

func TestSanitizeMessages(t *testing.T) {
	testFile, err := os.CreateTemp(os.TempDir(), "hurabotTestSanitizeMessages*.csv")

	if err != nil {
		t.Fatal(err)
	}

	testCsvData := `ID,Timestamp,Contents,Attachments
000000000000000001,2000-01-01 12:00:00.000000+00:00,Test message #1,
000000000000000002,2000-01-01 12:00:00.000000+00:00,Test message #2,
000000000000000003,2000-01-01 12:00:00.000000+00:00,Attachment message,https://test.attachment.test/
000000000000000004,2000-01-01 12:00:00.000000+00:00,https://test.attachment.test/,
000000000000000005,2000-01-01 12:00:00.000000+00:00,Wow animated emoji <a:123456789>,
000000000000000006,2000-01-01 12:00:00.000000+00:00,Hey <@123456789> this is a mention!,
000000000000000007,2000-01-01 12:00:00.000000+00:00,<#123456789> is a pretty cool channel,
`

	if _, err := testFile.WriteString(testCsvData); err != nil {
		t.Fatal(err)
	}

	if err := testFile.Close(); err != nil {
		t.Errorf("failed to close test file: %v", err)
	}

	testFile, err = os.Open(testFile.Name())

	if err != nil {
		t.Fatal(err)
	}

	parsedMessages, err := ProcessMessagesCSV(testFile)

	if err != nil {
		t.Errorf("failed to parse messages.csv: %v", err)
	}

	sanitizedMessages := SanitizeMessages(parsedMessages)

	if len(sanitizedMessages) != 21 && t.Failed() == false {
		t.Errorf("error sanitizing messages: sanitized messages length was %d instead of 21", err)
	}

	if err := testFile.Close(); err != nil {
		t.Logf("failed to close test file %s: %v", testFile.Name(), err)
	}

	if err := os.Remove(testFile.Name()); err != nil {
		t.Logf("failed to remove test file %s: %v", testFile.Name(), err)
	}
}
