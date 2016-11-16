package main

import (
	"bytes"
	"net/http"
	"os"
	"strings"
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

func TestFilterReader(t *testing.T) {
	var filterTests = []struct {
		content  string
		filter   string
		expected []string
	}{
		{`And the Lord spake, saying, 'First shalt thou take out the Holy Pin.  Then,
		  shalt thou count to three, no more, no less. Three shall be the number`,
			"Holy Pin",
			[]string{"Holy Pin"},
		},
		{`<td><a href="updates.20161031.2355.gz">updates.20161031.2355.gz</a></td>`,
			"updates.\\d{8}",
			[]string{"updates.20161031", "updates.20161031"},
		},
	}

	for _, test := range filterTests {
		// write test.content into buffer
		var testBuffer bytes.Buffer
		testBuffer.WriteString(test.content)

		// filter buffer for test.filter
		result, err := filterReader(test.filter, bytes.NewReader(testBuffer.Bytes()))
		if err != nil {
			t.Error("Failed to read %v as []byte", test.content)
		}

		// test the results
		resultString := strings.Join(result, ",")
		expectedString := strings.Join(test.expected, ",")
		if resultString != expectedString {
			t.Error("Got:", resultString, "Expected:", expectedString)
		}
	}
}

func TestPickWhichStrings(t *testing.T) {
	stringsToTest := []string{"First", "Second", "Third", "Last"}
	var pickTests = []struct {
		which      string
		stringList []string
		expected   []string
	}{
		{"first", stringsToTest, []string{stringsToTest[0]}},
		{"FIRST", stringsToTest, []string{stringsToTest[0]}},
		{"last", stringsToTest, []string{stringsToTest[len(stringsToTest)-1]}},
		{"LAST", stringsToTest, []string{stringsToTest[len(stringsToTest)-1]}},
		{"all", stringsToTest, stringsToTest},
		{"ALL", stringsToTest, stringsToTest},
		{"some", stringsToTest, []string{""}}, // invalid which string
	}

	for _, test := range pickTests {
		result := pickWhichStrings(test.stringList, test.which)
		// test length
		if len(result) != len(test.expected) {
			t.Error("Got length:", len(result), "Expected Length:", len(test.expected))
		}

		// test content
		resultString := strings.Join(result, ",")
		expectedString := strings.Join(test.expected, ",")
		if resultString != expectedString {
			t.Error("Got:", resultString, "Expected:", expectedString)
		}
	}
}

func TestUniqueStrings(t *testing.T) {
	var tests = []struct {
		stringList []string
		uniques    []string
	}{
		{[]string{"one", "two", "three"}, []string{"one", "two", "three"}},
		{[]string{"one", "one", "one"}, []string{"one"}},
	}

	for _, test := range tests {
		resultString := strings.Join(uniqueStrings(test.stringList), ",")
		uniquesString := strings.Join(test.uniques, ",")
		if resultString != uniquesString {
			t.Error("Got:", resultString, "Expected:", uniquesString)
		}
	}
}
