package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestConfigLoadConfig(t *testing.T) {
	testBotConfig := MainBotConfig{
		AuthenticationToken: "testToken",
		GuildID:             "12345",
		ModelDirectory:      "logDirectory",
		ModelsToUse:         nil,
		MaxWords:            123,
		LogDir:              "logDirectory",
		LogLevel:            "default",
	}

	testFile, err := os.CreateTemp(os.TempDir(), "hurabotConfigTestFile")

	if err != nil {
		t.Fatal(err)
	}

	enc := json.NewEncoder(testFile)
	enc.SetIndent("", "\t")
	err = enc.Encode(testBotConfig)

	if err != nil {
		t.Fatal(err)
	}

	err = ConfigLoadConfig(testFile)

	if err != nil {
		t.Errorf("failed to load test config: %v", err)
	}

	if LoadedConfig.AuthenticationToken != "testToken" {
		t.Errorf("loaded config doesn't match test config")
	}

	if err := testFile.Close(); err != nil {
		t.Logf("failed to close test file %s: %v", testFile.Name(), err)
	}

	if err := os.Remove(testFile.Name()); err != nil {
		t.Logf("failed to remove test file %s: %v", testFile.Name(), err)
	}

	if err != nil {
		t.Errorf("failed to remove test file: %v", err)
	}
}
