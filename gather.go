// Gather will download files over http; it requires a JSON config file to
// direct it's activities.
package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pladdy/lumberjack"
)

func main() {
	lumberjack.StartLogging()
	processStart := time.Now()

	if len(os.Args) == 1 {
		lumberjack.Fatal("Config file is required as an argument")
	}

	lumberjack.Info("Reading in config " + os.Args[1])
	config := unmarshalConfig(os.Args[1], time.Now())

	if config.isValid() != true {
		lumberjack.Error("Invalid config.  Update config according to GatherConfig struct.")
		printDocs()
		os.Exit(1)
	}

	filesToDownload := filesToDownload(config)
	lumberjack.Info("Files to download: %v", filesToDownload)

	downloadFiles(config, filesToDownload)
	lumberjack.Info("Download completed in %v", time.Since(processStart))
}

// Given a config and a list of files, download the files using config settings
// to set file name
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

// Given the gather config, return the files on a host (scrape) or return the
// uri (simple config), to download
func filesToDownload(config GatherConfig) []string {
	if config.FileListHost != "" {
		response, err := http.Get(config.FileListHost)
		if err != nil {
			lumberjack.Panic("Failed to get list of files: %v", err)
		}
		defer response.Body.Close()

		matchingFiles, err := filterReader(config.FilePattern, response.Body)
		if err != nil {
			lumberjack.Panic("Failed to find matching files: %v", err)
		}

		matchingFiles = uniqueStrings(matchingFiles)
		sort.Strings(matchingFiles)
		matchingFiles = pickWhichStrings(matchingFiles, config.WhichFilesToGet)

		// prepend the root path to the files scraped
		for i, file := range matchingFiles {
			matchingFiles[i] = config.FileListHost + "/" + file
		}
		return matchingFiles
	} else {
		return []string{config.DownloadHost}
	}
}

// Given a string and io.ReadCloser, scan through the reader and return any
// strings that match the filter
func filterReader(filter string, rc io.Reader) ([]string, error) {
	pattern := regexp.MustCompile(filter)
	scanner := bufio.NewScanner(rc)
	var matches []string

	for scanner.Scan() {
		// find all string matches
		matchedStrings := pattern.FindAllString(scanner.Text(), -1)

		if len(matchedStrings) > 0 {
			lumberjack.Debug("Matches found: %v", matchedStrings)
			for _, match := range matchedStrings {
				matches = append(matches, match)
			}
		}
	}

	return matches, scanner.Err()
}

// Given a list of strings, pick the one(s) you want based on named quantity
//    "all"    -> give me all of them
//    "first"  -> give me the first one after sorting
//    "last" -> give me the last one after sorting
func pickWhichStrings(stringList []string, which string) (picked []string) {
	lumberjack.Debug("Which to pick: %v", which)

	if len(stringList) > 0 {
		switch strings.ToLower(which) {
		case "all":
			picked = stringList
		case "first":
			picked = append(picked, stringList[0])
		case "last":
			picked = append(picked, stringList[len(stringList)-1])
		default:
			picked = append(picked, "")
		}
	}

	lumberjack.Debug("Picked string(s): %v", picked)
	return
}

// Print godoc for gather
func printDocs() {
	command := exec.Command("godoc", "-src", "github.com/pladdy/gather")
	output, err := command.Output()
	if err != nil {
		os.Stdout.WriteString(fmt.Sprintf("Unable to print docs: %v", err))
	}

	os.Stderr.WriteString(fmt.Sprintf("%s", output))
}

// Given a ContentLength and a file peridically log how much is downloaded
func trackDownload(contentLength int64, filePath string) {
	if contentLength == -1 {
		lumberjack.Info("Content-Length not available, can't track download")
		return
	}

	var fileSize int64 = 0
	const durationToSleep int64 = 10

	for fileSize < contentLength {
		time.Sleep(time.Second * time.Duration(durationToSleep))

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

func uniqueStrings(stringList []string) (uniques []string) {
	lumberjack.Debug("List: %v", stringList)
	seenStrings := make(map[string]bool)

	for _, item := range stringList {
		if seenStrings[item] != true {
			uniques = append(uniques, item)
			seenStrings[item] = true
		}
	}
	lumberjack.Debug("Unique List: %v", uniques)
	return
}
