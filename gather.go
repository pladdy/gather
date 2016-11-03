// Gather will download files over http; it requires a JSON config file to
// direct it's activities.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
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

	// Simple download
	DownloadHost string // Host to download from

	// Scrape download
	FileListHost string // Host to scrape files from
	DownloadRoot string // For scraping, the root to use for matching files
	FilePattern  string // Pattern to look for when scraping for files
	FilesToGet   string // 'all' or 'latest'; needed for scraping
}

func main() {
	lumberjack.StartLogging()
	processStart := time.Now()

	if len(os.Args) == 1 {
		lumberjack.Fatal("Config file is required as an argument")
	}

	lumberjack.Info("Reading in config " + os.Args[1])
	config := unmarshalGatherConfig(os.Args[1], time.Now())

	if config.isValid() != true {
		lumberjack.Error("Invalid config settings, update config according to GatherConfig struct.")
		printDocs()
		os.Exit(1)
	}

	filesToDownload := filesToDownload(config)
	lumberjack.Info("Files to download: %v", filesToDownload)

	downloadFiles(config, filesToDownload)
	lumberjack.Info("Download completed in %v", time.Since(processStart))
}

// Given a config and a list of files, download them
func downloadFiles(config GatherConfig, remoteFiles []string) {
	for _, remoteFile := range remoteFiles {
		parts := strings.Split(remoteFile, "/")
		fileName := parts[len(parts)-1]
		path := "./" + config.DestinationFilePrefix + fileName
		downloadFile(remoteFile, path)
	}
}

// Given a uri and a file path, get the data and save to the path
func downloadFile(downloadLink string, path string) error {
	lumberjack.Info("Downloading %v to %v", downloadLink, path)

	response, err := http.Get(downloadLink)
	if err != nil {
		lumberjack.Panic("Error Getting %v: %v", downloadLink, err)
	}
	defer response.Body.Close()

	downloadHandle, err := os.Create(path)
	if err != nil {
		lumberjack.Panic("Error creating file %v", err)
	}
	defer downloadHandle.Close()

	go trackDownload(response.ContentLength, path)

	_, err = io.Copy(downloadHandle, response.Body)
	if err != nil {
		lumberjack.Panic("Failed to download file %v", downloadLink)
	}

	lumberjack.Info("Finished downloading %v", downloadLink)
	return err
}

// Given the config, reach out to the file list host and identify what files to
// download
func filesToDownload(config GatherConfig) (filesToDownload []string) {
	if config.FileListHost != "" {
		remoteList, err := http.Get(config.FileListHost)
		if err != nil {
			lumberjack.Panic("Failed to get list of files: %v", err)
		}

		matchingFiles, err := findMatches(config.FilePattern, remoteList)
		if err != nil {
			lumberjack.Panic("Failed to find matching files: %v", err)
		}

		// prepend the root path to the files scraped
		for i, file := range matchingFiles {
			matchingFiles[i] = config.DownloadRoot + "/" + file
		}
		return pickFilesToGet(matchingFiles, config.FilesToGet)
	} else {
		return []string{config.DownloadHost}
	}
}

// Given a string to match and an http response, return matches from response
func findMatches(toFind string, response *http.Response) ([]string, error) {
	defer response.Body.Close()

	fileRegexp := regexp.MustCompile(toFind)
	scanner := bufio.NewScanner(response.Body)
	var matches []string

	for scanner.Scan() {
		match := fileRegexp.FindAllStringSubmatch(scanner.Text(), 1)
		if len(match) > 0 {
			lumberjack.Debug("Match found: %v", match)
			matches = append(matches, match[0][1])
		}
	}

	return matches, scanner.Err()
}

// Given a list of file and a directive for which ones to return, return the
// files asked for
func pickFilesToGet(files []string, filesToGet string) []string {
	var pickedFiles []string

	if len(files) > 0 {
		switch filesToGet {
		case "latest":
			sort.Strings(files)
			pickedFiles = append(pickedFiles, files[len(files)-1])
		case "all":
			pickedFiles = files
		default:
			panic("Config value for 'FilesToGet' should be 'latest' or 'all'")
		}
	}

	lumberjack.Debug("Picked file(s): %v", pickedFiles)
	return pickedFiles
}

// Print godoc for gather
func printDocs() {
	command := exec.Command("godoc", "-src", "github.com/pladdy/gather")
	output, err := command.Output()
	if err != nil {
		os.Stdout.Write([]byte(fmt.Sprintf("Unable to print docs: %v", err)))
	}

	os.Stderr.Write([]byte(fmt.Sprintf("%s", output)))
}

// Given a ContentLength, a file, and a logger, peridically log how much is
// downloaded
func trackDownload(contentLength int64, filePath string) {
	if contentLength == -1 {
		lumberjack.Info("Content-Length not available, can't track download")
		return
	}

	var fileSize int64 = 0
	const TimeToSleep int64 = 15

	for fileSize < contentLength {
		time.Sleep(time.Second * time.Duration(TimeToSleep))

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			lumberjack.Error("Couldn't get info on file %v to track", filePath)
			return
		}

		fileSize = fileInfo.Size()
		progress := float64(fileSize) / float64(contentLength) * 100
		lumberjack.Info("Download progress: %.2f%%", progress)
	}
}

// Given a JSON config file name, read in the file and return a go lang
// data structure
func unmarshalGatherConfig(fileName string, theTime time.Time) GatherConfig {
	contents, err := ioutil.ReadFile(fileName)
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

// Given a GatherCOnfig, validate it
func (config *GatherConfig) isValid() bool {
	itIsValid := false

	if config.DownloadHost != "" {
		itIsValid = true
	}

	if config.FileListHost != "" &&
		config.DownloadRoot != "" &&
		config.FilePattern != "" &&
		config.FilesToGet != "" {
		itIsValid = true
	}

	return itIsValid
}
