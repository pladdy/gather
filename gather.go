package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/pladdy/lumberjack"
	"github.com/pladdy/timepiece"
)

func main() {
	lumberjack.StartLogging()

	if len(os.Args) == 1 {
		lumberjack.Fatal("Config file is required as an argument")
	}

	config := readConfig(os.Args[1], time.Now())
	var filesToDownload []string

	if config["file list host"] != nil {
		matchingFiles, err := getFileList(config)
		if err != nil {
			lumberjack.Fatal("Failed to get file list")
		}

		if len(matchingFiles) == 0 {
			lumberjack.Info("No files found to download")
			os.Exit(0)
		}

		switch config["files to get"] {
		case "latest":
			lumberjack.Info("Getting latest file from matches")
			sort.Strings(matchingFiles)
			filesToDownload =
				append(filesToDownload, matchingFiles[len(matchingFiles)-1])
			lumberjack.Info("Latest file is %v", filesToDownload[0])
		case "all":
			lumberjack.Info("Getting all matching files")
			filesToDownload = matchingFiles
			lumberjack.Info("Files to download: %v", filesToDownload)
		default:
			lumberjack.Panic("Don't know which files to get; update config with 'latest' or 'all'")
		}

	} else {
		filesToDownload =
			append(filesToDownload, fmt.Sprintf("%v", config["download host"]))
	}

	if len(filesToDownload) == 0 {
		lumberjack.Info("No files to download")
	}

	// Get files
	for _, file := range filesToDownload {
		downloadLink := fmt.Sprintf("%v", config["download root"]) + "/" + file
		filePath := "./" + file
		gatherData(downloadLink, filePath)
	}
}

/* private */

// Given a uri and a file path, get the data and save to the path
func gatherData(downloadLink string, filePath string) error {
	lumberjack.Info("Getting response from " + downloadLink)

	response, err := http.Get(downloadLink)
	if err != nil {
		lumberjack.Panic("Error in Get: %v", err)
	}
	defer response.Body.Close()

	// Download file
	lumberjack.Info("Starting to download " + downloadLink)

	downloadHandle, err := os.Create(filePath)
	if err != nil {
		lumberjack.Panic("Error creating file %v", err)
	}
	defer downloadHandle.Close()

	_, err = io.Copy(downloadHandle, response.Body)
	if err != nil {
		lumberjack.Error("Failed to download file %v", downloadLink)
	}

	lumberjack.Info("Finished downloading %v", downloadLink)
	return err
}

func getFileList(config map[string]interface{}) (files []string, err error) {
	// Get response from host
	lumberjack.Info("Calling Get on %v", config["file list host"])

	response, err := http.Get(fmt.Sprintf("%v", config["file list host"]))
	if err != nil {
		lumberjack.Panic("Error in Get: ", err)
	}
	defer response.Body.Close()

	// Parse response
	fileRegexp := regexp.MustCompile(fmt.Sprintf("%v", config["file pattern"]))
	scanner := bufio.NewScanner(response.Body)
	var matchingFiles []string

	for scanner.Scan() {
		match := fileRegexp.FindAllStringSubmatch(scanner.Text(), 1)
		if len(match) > 0 {
			lumberjack.Info("Found match: %v", match[0])
			matchingFiles = append(matchingFiles, match[0][1])
		}
	}
	if err := scanner.Err(); err != nil {
		lumberjack.Panic("Error in scanning response: ", err)
	}

	return matchingFiles, err
}

// GIven a JSON config file name, read in the file and return a go lang
// data structure
func readConfig(fileName string, theTime time.Time) map[string]interface{} {
	lumberjack.Info("Reading in config " + fileName)

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		lumberjack.Fatal("Couldn't read your config file.  Contents: %v", contents)
	}

	lumberjack.Info("Replacing time vars in config")

	replacedContents :=
		timepiece.ReplaceTime(string(contents), timepiece.TimeToTimePiece(theTime))

	lumberjack.Info("Unmarshalling config from JSON")

	var configJson interface{}
	err = json.Unmarshal([]byte(replacedContents), &configJson)
	if err != nil {
		lumberjack.Fatal("Failed to Unmarshal JSON; contents: %v", configJson)
	}

	return configJson.(map[string]interface{})
}
