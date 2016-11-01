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
	DownloadName string // Name to be given to download file

	// Scrape download
	FileListHost string // Host to scrape files from
	DownloadRoot string // For scraping, the root to use for matching files
	FilePattern  string // Pattern to look for when scraping for files
	FilesToGet   string // 'all' or 'latest'; needed for scraping
}

func main() {
	logger := lumberjack.New()
	processStart := time.Now()

	if len(os.Args) == 1 {
		logger.Fatal("Config file is required as an argument")
	}

	logger.Info("Reading in config " + os.Args[1])
	config := unmarshalGatherConfig(os.Args[1], time.Now())

	if config.isValid() != true {
		logger.Error("Invalid config settings, update config according to GatherConfig struct.")
		printDocs()
		os.Exit(1)
	}

	filesToDownload := filesToDownload(config, logger)
	logger.Info("Files to download: %v", filesToDownload)

	if len(filesToDownload) == 0 && config.FileListHost != "" {
		logger.Info("No files to download")
		os.Exit(0)
	}

	downloadFiles(config, filesToDownload, logger)
	logger.Info("Download completed in %v", time.Since(processStart))
}

// Given a config and a list of files, download them
func downloadFiles(config GatherConfig, files []string, logger lumberjack.Logger) {
	if files == nil {
		downloadFile(config.DownloadHost, config.DownloadName, logger)
	} else {
		for _, file := range files {
			downloadLink := config.DownloadRoot + "/" + file

			path := "./" + config.DestinationFilePrefix + file
			logger.Info("Downloading %v", path)
			downloadFile(downloadLink, path, logger)
		}
	}
}

// Given a uri and a file path, get the data and save to the path
func downloadFile(downloadLink string, path string, logger lumberjack.Logger) error {
	response, err := http.Get(downloadLink)
	if err != nil {
		logger.Panic("Error Getting %v: %v", downloadLink, err)
	}
	defer response.Body.Close()

	downloadHandle, err := os.Create(path)
	if err != nil {
		logger.Panic("Error creating file %v", err)
	}
	defer downloadHandle.Close()

	go trackDownload(response.ContentLength, path, logger)

	_, err = io.Copy(downloadHandle, response.Body)
	if err != nil {
		logger.Error("Failed to download file %v", downloadLink)
	}

	logger.Info("Finished downloading %v", downloadLink)
	return err
}

// Given the config, reach out to the file list host and identify what files to
// download
func filesToDownload(config GatherConfig, logger lumberjack.Logger) []string {
	var filesToDownload []string

	if config.FileListHost != "" {
		remoteFiles, err := http.Get(config.FileListHost)
		if err != nil {
			logger.Panic("Failed to get list of files: %v", err)
		}
		matchingFiles, err := findMatchingFiles(config.FilePattern, remoteFiles)
		if err != nil {
			logger.Panic("Failed to find matching files: %v", err)
		}
		filesToDownload = pickFilesToGet(matchingFiles, config.FilesToGet)
	}

	return filesToDownload
}

// Given a string to match and an http response, return matching files from response
func findMatchingFiles(toFind string, response *http.Response) ([]string, error) {
	defer response.Body.Close()

	fileRegexp := regexp.MustCompile(toFind)
	scanner := bufio.NewScanner(response.Body)
	var matchingFiles []string

	for scanner.Scan() {
		match := fileRegexp.FindAllStringSubmatch(scanner.Text(), 1)
		if len(match) > 0 {
			matchingFiles = append(matchingFiles, match[0][1])
		}
	}

	return matchingFiles, scanner.Err()
}

// Given a list of file and a directive for which ones to return, return the
// files asked for
func pickFilesToGet(files []string, filesToGet string) []string {
	var pickedFiles []string

	if len(files) > 0 {
		switch filesToGet {
		case "latest":
			sort.Strings(files)
			pickedFiles =
				append(pickedFiles, files[len(files)-1])
		case "all":
			pickedFiles = files
		default:
			panic("Don't know which files to pick; config 'FilesToGet' should be 'latest' or 'all'")
		}
	}

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
func trackDownload(contentLength int64, filePath string, logger lumberjack.Logger) {
	if contentLength == -1 {
		logger.Warn("Content-Length not available, can't track download")
		return
	}

	var fileSize int64 = 0
	const TimeToSleep int64 = 15

	for fileSize < contentLength {
		time.Sleep(time.Second * time.Duration(TimeToSleep))

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logger.Warn("Couldn't get info on file %v; not tracking", filePath)
		}

		fileSize = fileInfo.Size()
		progress := float64(fileSize) / float64(contentLength) * 100
		logger.Info("Download progress: %.2f%%", progress)
	}
}

// Given a JSON config file name, read in the file and return a go lang
// data structure
func unmarshalGatherConfig(fileName string, theTime time.Time) GatherConfig {
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic("Couldn't read config file.  Contents: " + string(contents))
	}

	replacedContents :=
		timepiece.ReplaceTime(string(contents), timepiece.TimeToTimePiece(theTime))

	var config GatherConfig
	err = json.Unmarshal([]byte(replacedContents), &config)
	if err != nil {
		panic("Failed to Unmarshal JSON")
	}

	return config
}

// Given a GatherCOnfig, validate it
func (config *GatherConfig) isValid() bool {
	itIsValid := false

	if config.DownloadHost != "" && config.DownloadName != "" {
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
