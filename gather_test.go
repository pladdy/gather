package main

import "testing"

func TestIsValid(t *testing.T) {
	var simpleConfig = GatherConfig{
		DownloadHost: "some-host",
		DownloadName: "some-name",
	}

	if simpleConfig.isValid() != true {
		t.Error("Expected", simpleConfig, "to be true")
	}
}
