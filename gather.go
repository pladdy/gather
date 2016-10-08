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
	lumberjack.StartLogging()

	if len(os.Args) == 1 {
		lumberjack.Fatal("Config file is required as an argument")
	}

	config := unmarshalGatherConfig(os.Args[1], time.Now())

	if config.isValid() != true {
		lumberjack.Error("Invalid config settings, update config according to GatherConfig struct.")
		printDocs()
		os.Exit(1)
	}

	filesToDownload := filesToDownload(config)

	if len(filesToDownload) == 0 && config.FileListHost != "" {
		lumberjack.Info("No files to download")
		os.Exit(0)
	}

	downloadFiles(config, filesToDownload)
}

// Given a config and a list of files, download them
func downloadFiles(config GatherConfig, filesToDownload []string) {
	if filesToDownload == nil {
		downloadFile(config.DownloadHost, config.DownloadName)
	} else {
		for _, file := range filesToDownload {
			downloadLink := config.DownloadRoot + "/" + file
			filePath := "./" + file
			downloadFile(downloadLink, filePath)
		}
	}
}

// Given a uri and a file path, get the data and save to the path
func downloadFile(downloadLink string, filePath string) error {
	lumberjack.Info("Getting response from " + downloadLink)

	response, err := http.Get(downloadLink)
	if err != nil {
		lumberjack.Panic("Error Getting %v: %v", downloadLink, err)
	}
	defer response.Body.Close()

	lumberjack.Info("Downloading " + downloadLink)

	downloadHandle, err := os.Create(filePath)
	if err != nil {
		lumberjack.Panic("Error creating file %v", err)
	}
	defer downloadHandle.Close()

	go trackDownload(response.ContentLength, filePath)

	_, err = io.Copy(downloadHandle, response.Body)
	if err != nil {
		lumberjack.Error("Failed to download file %v", downloadLink)
	}

	lumberjack.Info("Finished downloading %v", downloadLink)
	return err
}

func fileList(config GatherConfig) *http.Response {
	lumberjack.Info("Calling Get on %v", config.FileListHost)

	response, err := http.Get(config.FileListHost)
	if err != nil {
		lumberjack.Panic("Error in Get: ", err)
	}

	return response
}

// Given the config, reach out to the file list host and identify what files to
// download
func filesToDownload(config GatherConfig) []string {
	var filesToDownload []string

	if config.FileListHost != "" {
		matchingFiles := findMatchingFiles(config.FilePattern, fileList(config))
		filesToDownload = pickFilesToGet(matchingFiles, config.FilesToGet)
	}

	lumberjack.Info("Files to download: %v", filesToDownload)
	return filesToDownload
}

// Given a string to match and an http response, return matching files from response
func findMatchingFiles(stringToFind string, response *http.Response) []string {
	defer response.Body.Close()

	fileRegexp := regexp.MustCompile(stringToFind)
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

	return matchingFiles
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
			lumberjack.Panic("Don't know which 'files to get'; update config with key set to 'latest' or 'all'")
		}
	}

	return pickedFiles
}

// Print godoc for gather
func printDocs() {
	command := exec.Command("godoc", "-src", "github.com/pladdy/gather")
	output, err := command.Output()
	if err != nil {
		lumberjack.Warn("I tried to print out the godoc docs but it didn't work.  Sorry.")
	}

	lumberjack.Error(fmt.Sprintf("%s", output))
}

// Given a ContentLength and a file, peridically log how much is downloaded
func trackDownload(contentLength int64, filePath string) {
	if contentLength == -1 {
		lumberjack.Warn("Content-Length not available, can't track download")
		return
	}

	var fileSize int64 = 0
	const TimeToSleep int64 = 15

	for fileSize < contentLength {
		time.Sleep(time.Second * time.Duration(TimeToSleep))

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			lumberjack.Warn("Couldn't get info on file %v; not tracking", filePath)
		}

		fileSize = fileInfo.Size()
		progress := float64(fileSize) / float64(contentLength) * 100
		lumberjack.Info("Download progress: %.2f%%", progress)
	}

	return
}

// Given a JSON config file name, read in the file and return a go lang
// data structure
func unmarshalGatherConfig(fileName string, theTime time.Time) GatherConfig {
	lumberjack.Info("Reading in config " + fileName)

	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		lumberjack.Fatal("Couldn't read config file.  Contents: %v", contents)
	}

	lumberjack.Info("Replacing time vars in config")

	replacedContents :=
		timepiece.ReplaceTime(string(contents), timepiece.TimeToTimePiece(theTime))

	lumberjack.Info("Unmarshalling config from JSON")

	var config GatherConfig
	err = json.Unmarshal([]byte(replacedContents), &config)
	if err != nil {
		lumberjack.Fatal("Failed to Unmarshal JSON; contents: %v", config)
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
