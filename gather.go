// Gather provides a CLI for downloading content from remote locations.
package main

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/pladdy/lumberjack"
)

// Create a global parser that can have commands/options added to it.  See
// cli.go for more information.
var commonOptions CommonOptions
var parser = flags.NewParser(&commonOptions, flags.Default)

// These options are global as well but commands are added during initalization.
// See cli.go for more details
var downloadOptions DownloadOptions
var scrapeOptions ScrapeOptions

func main() {
	lumberjack.StartLogging()
	processStart := time.Now()

	// Parse commandl line; if something goes wrong in flags we verify tests
	// aren't being run before exiting with non-zero exit.
	_, err := parser.Parse()

	if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
		os.Exit(0)
	} else if flagsErr != nil && isTesting(os.Args[0]) == false {
		os.Exit(1)
	}

	// Perform command
	switch commandToRun(downloadOptions, scrapeOptions) {
	case "scrape":
		files := filesToScrape(scrapeOptions)
		lumberjack.Info("Files to scrape: %v", files)
		downloadFiles(commonOptions.SaveAs, files)
	case "download":
		lumberjack.Info("File to download: %v", downloadOptions.URI)
		downloadFiles(commonOptions.SaveAs, []string{downloadOptions.URI})
	}

	lumberjack.Info("Process completed in %v", time.Since(processStart))
}

// Given the CLI options and a list of files, download the files
func downloadFiles(saveAs string, uris []string) {
	i := 0
	for _, uri := range uris {
		path := saveAs
		if len(uris) > 1 {
			s := strconv.Itoa(i)
			path = path + "_" + s
		}
		downloadFile(uri, path)
		i += 1
	}
}

// Given a uri and a file path, get the data and save to the path
func downloadFile(uri string, path string) error {
	lumberjack.Info("Downloading %v to %v", uri, path)

	response, err := http.Get(uri)
	if err != nil {
		lumberjack.Panic("Error Getting %v: %v", uri, err)
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
		lumberjack.Panic("Failed to download file %v", uri)
	}

	lumberjack.Info("Finished downloading %v", uri)
	return err
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
