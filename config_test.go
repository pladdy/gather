package main

import (
	"testing"
	"time"
)

// set up some configs
var badConfig = GatherConfig{
	DownloadHost: "some-host",
}

var simpleConfig = GatherConfig{
	Type:         "Simple",
	DownloadHost: "some-host",
}

var scrapeConfig = GatherConfig{
	Type:            "Scrape",
	DownloadHost:    "some-host",
	FileListHost:    "http://some.file.host",
	FilePattern:     "*.txt",
	WhichFilesToGet: "last",
}

// run tests

func TestIsValid(t *testing.T) {
	var validationTests = []struct {
		config        GatherConfig
		shouldBeValid bool
	}{
		{simpleConfig, true},
		{scrapeConfig, true},
		{badConfig, false},
	}

	for _, test := range validationTests {
		result := test.config.isValid()
		if result != test.shouldBeValid {
			t.Error(
				"Expected",
				test.config, "to be",
				test.shouldBeValid,
				"Got:",
				result)
		}
	}
}

func TestValidScrapeConfig(t *testing.T) {
	if validScrapeConfig(&scrapeConfig) != true {
		t.Error("Expected", scrapeConfig, "to be true")
	}
}

func TestValidSimpleConfig(t *testing.T) {
	if validSimpleConfig(&simpleConfig) != true {
		t.Error("Expected", simpleConfig, "to be true")
	}
}

func TestUnmarshalConfig(t *testing.T) {
	config := unmarshalConfig("testdata/simple.json", time.Now())

	if config.DownloadHost == "" {
		t.Error("DownloadHost should not be empty")
	}
}
