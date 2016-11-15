package main

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pladdy/lumberjack"
	"github.com/pladdy/timepiece"
)

// A gather config has settings that dictate what hosts have data pulled
// from, the files that should be gathered, etc.
//
// Keys grouped together are required to be set in the config
type GatherConfig struct {
	DestinationFilePrefix string
	Type                  string

	// Simple download
	DownloadHost string // Host to download from

	// Scrape download
	FileListHost    string // Host to scrape files from
	FilePattern     string // Pattern to look for when scraping for files
	WhichFilesToGet string // 'all' or 'latest'; needed for scraping
}

var GatherConfigTypes []string = []string{
	"Simple",
	"Scrape",
}

// Given a JSON config file name, read in the file and return a go lang
// data structure
func unmarshalConfig(path string, theTime time.Time) GatherConfig {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		lumberjack.Panic("Couldn't read config file.: %v", err)
	}

	contentsString := string(contents)
	lumberjack.Debug("Contents of config:\n" + contentsString)

	replacedContents :=
		timepiece.ReplaceTime(contentsString, timepiece.TimeToTimePiece(theTime))

	if contentsString != replacedContents {
		lumberjack.Debug("New contents:\n" + replacedContents)
	}

	var config GatherConfig
	err = json.Unmarshal([]byte(replacedContents), &config)
	if err != nil {
		lumberjack.Panic("Failed to Unmarshal JSON: %v", err)
	}

	return config
}

func validScrapeConfig(config *GatherConfig) bool {
	if config.FileListHost != "" &&
		config.FilePattern != "" &&
		config.WhichFilesToGet != "" {
		return true
	}
	return false
}

func validSimpleConfig(config *GatherConfig) bool {
	if config.DownloadHost != "" {
		return true
	}
	return false
}

// Given a GatherCOnfig, validate it
func (config *GatherConfig) isValid() bool {
	itIsValid := false

	switch config.Type {
	case "Simple":
		itIsValid = validSimpleConfig(config)
	case "Scrape":
		itIsValid = validScrapeConfig(config)
	}

	return itIsValid
}
