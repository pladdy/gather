package main

import (
	"fmt"
	"os"
	"regexp"
)

// Create option structs
type CommonOptions struct {
	SaveAs string `short:"s" long:"save-as" description:"Path to save downloads to" required:"true"`
}

type DownloadOptions struct {
	URI string `short:"u" long:"uri" description:"Host to download from" required:"true"`
}

type ScrapeOptions struct {
	URI             string `short:"u" long:"uri" description:"Host to scrape files from" required:"true"`
	Pattern         string `short:"p" long:"pattern" description:"Pattern to look for when scraping for files" required:"true"`
	WhichFilesToGet string `short:"w" long:"which" description:"Which files to get: 'all' or 'latest'" required:"true"`
}

// init depends on a global parser to add commands to.  It should be declared in
// the main package
func init() {
	// Create download command
	_, err := parser.AddCommand(
		"download",
		"Download a URL contents to file",
		"Given a URL, download it's contents to a file",
		&downloadOptions)

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}

	// Create scrape command
	_, err = parser.AddCommand(
		"scrape",
		"Scrape a URL for files",
		"Scrape a URL for file patterns and download matching files",
		&scrapeOptions)

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

// Identify what command was given...there has to be a better way to do this
// TODO: find a better way...
func commandToRun(dl DownloadOptions, sc ScrapeOptions) string {
	if dl.URI != "" {
		return "download"
	} else if sc.URI != "" {
		return "scrape"
	}
	return ""
}

// Check first arg and see if it ends in .test; if so tests are being run
func isTesting(firstArg string) bool {
	matched, _ := regexp.MatchString(".test$", firstArg)
	return matched
}

func synopsis() string {
	return "Synopsis:\n  gather is a CLI for downloading URIs."
}
