package main

import (
	"bufio"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/pladdy/lumberjack"
)

// Given the scrape options, find files to scrape that match the options
// provided.
func filesToScrape(opts ScrapeOptions) []string {
	response, err := http.Get(opts.URI)
	if err != nil {
		lumberjack.Panic("Failed to get list of files: %v", err)
	}
	defer response.Body.Close()

	files, err := filterReader(opts.Pattern, response.Body)
	if err != nil {
		lumberjack.Panic("Failed to find matching files: %v", err)
	}

	sort.Strings(files)
	files = uniqueStrings(files)
	files = pickWhichStrings(files, opts.WhichFilesToGet)

	// prepend the root path to the files scraped
	for i, file := range files {
		files[i] = opts.URI + "/" + file
	}
	return files
}

// Given a filter string and io.Reader, scan through the reader and return any
// strings that match the filter
func filterReader(filter string, r io.Reader) ([]string, error) {
	pattern := regexp.MustCompile(filter)
	scanner := bufio.NewScanner(r)
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
//    "last"   -> give me the last one after sorting
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
