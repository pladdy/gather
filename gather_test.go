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
