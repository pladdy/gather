package main

import "testing"

func TestCommandToRun(t *testing.T) {
	var tests = []struct {
		downloadOptions DownloadOptions
		scrapeOptions   ScrapeOptions
		expected        string
	}{
		{DownloadOptions{"some-uri"}, ScrapeOptions{}, "download"},
		{DownloadOptions{}, ScrapeOptions{"uri", "pattern", "all"}, "scrape"},
		{DownloadOptions{}, ScrapeOptions{}, ""},
	}

	for _, test := range tests {
		result := commandToRun(test.downloadOptions, test.scrapeOptions)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestIsTesting(t *testing.T) {
	var tests = []struct {
		firstArg string
		expected bool
	}{
		{"/T/go-build330525448/github.com/pladdy/gather/_test/gather.test", true},
		{"/nothing/to/test/here/gather.go", false},
		{"gather.test", true},
		{"gather.go", false},
	}

	for _, test := range tests {
		result := isTesting(test.firstArg)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestSynopsis(t *testing.T) {
	synopsis := synopsis()
	if synopsis == "" {
		t.Error("Got:", "", "Expected: a synopsis...")
	}
}
