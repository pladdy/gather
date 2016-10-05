package main

import (
	"bufio"
	"fmt"
	"io"
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

	// Read in config
	configContents := ReadConfig(os.Args[1])
	config := UnmarshalConfig(
		ReplaceTime(configContents, timepiece.TimeToTimePiece(time.Now())))

	// Get response from host
	lumberjack.Info("Calling Get from %v", config["download host"])

	response, err := http.Get(fmt.Sprintf("%v", config["download host"]))
	if err != nil {
		lumberjack.Panic("Error in Get: ", err)
	}
	defer response.Body.Close()

	// Check response
	if response.StatusCode != 200 {
		lumberjack.Fatal("Failed to get 200 status")
	}

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

	// Get latest file
	sort.Strings(matchingFiles)
	latestFile := matchingFiles[len(matchingFiles)-1]
	lumberjack.Info("Latest file is %v", latestFile)

	// Get file
	downloadLink := fmt.Sprintf("%v", config["download host"]) + "/" + latestFile
	lumberjack.Info("Calling Get from " + downloadLink)

	response, err = http.Get(downloadLink)
	if err != nil {
		lumberjack.Panic("Error in Get: %v", err)
	}
	defer response.Body.Close()

	// Download file
	lumberjack.Info("Starting download of " + downloadLink)

	downloadHandle, err := os.Create("./" + latestFile)
	if err != nil {
		lumberjack.Panic("Error creating file %v", err)
	}
	defer downloadHandle.Close()

	_, err = io.Copy(downloadHandle, response.Body)
	if err != nil {
		lumberjack.Error("Failed to download file %v", downloadLink)
	}

	lumberjack.Info("Finished downloading %v", downloadLink)
}
