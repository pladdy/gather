package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/pladdy/lumberjack"
)

func TestDownloadFile(t *testing.T) {
	// start logging and start a server
	lumberjack.Hush()
	go func() {
		http.ListenAndServe(":8080", http.FileServer(http.Dir("/tmp")))
	}()

	testDownloadPath := "./testDownload.txt"
	downloadFile("http://localhost:8080", testDownloadPath)

	// check to make sure content was downloaded
	fileInfo, _ := os.Stat(testDownloadPath)
	if fileInfo.Size() < 0 {
		t.Error("File size of download should be > 0")
	}

	// cleanup
	os.Remove(testDownloadPath)
}

func TestDownloadFiles(t *testing.T) {
	// start logging and start a server
	lumberjack.Hush()
	go func() {
		http.ListenAndServe(":8080", http.FileServer(http.Dir("/tmp")))
	}()

	// Multiple files
	uris := []string{"http://localhost:8080", "http://localhost:8080"}
	expectedDownloads := []string{"./test_0.html", "./test_1.html"}
	downloadFiles(uris, "./test.html")

	// check to make sure content was downloaded
	for _, file := range expectedDownloads {
		fileInfo, _ := os.Stat(file)
		if fileInfo.Size() < 0 {
			t.Error("File size of download should be > 0")
		}
		os.Remove(file)
	}

	// Single file
	uris = []string{"http://localhost:8080"}
	expectedDownloads = []string{"./test.html"}
	downloadFiles(uris, "./test.html")

	// check to make sure content was downloaded
	for _, file := range expectedDownloads {
		fileInfo, _ := os.Stat(file)
		if fileInfo.Size() < 0 {
			t.Error("File size of download should be > 0")
		}
		os.Remove(file)
	}
}

func TestIncrementPath(t *testing.T) {
	var tests = []struct {
		Path     string
		Inc      int
		Expected string
	}{
		{"/some/path/test.html", 1, "/some/path/test_1.html"},
		{"test.html", 1, "test_1.html"},
		{"test_html", 1, "test_html_1"},
	}

	for _, test := range tests {
		result := incrementPath(test.Path, test.Inc)
		if result != test.Expected {
			t.Error("Got:", result, "Expected:", test.Expected)
		}
	}
}
