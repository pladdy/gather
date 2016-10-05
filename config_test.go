package main

import "testing"

func TestUnmarshalConfig(t *testing.T) {
	configContents := ReadConfig("testdata/ripe_rrc_00.json")
	config := UnmarshalConfig(configContents)

	if config["test key"] != "test value" {
		t.Error("expected", "test value", "got", config["test key"])
	}
}
