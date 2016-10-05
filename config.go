package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/pladdy/lumberjack"
	"github.com/pladdy/timepiece"
)

// Read config file and return the bytes
func ReadConfig(configFile string) []byte {
	lumberjack.StartLogging()
	lumberjack.Info("Slurping config file " + configFile)

	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		lumberjack.Fatal(
			"Couldn't read your config file.  Contents: %v", configContents)
	}
	return configContents
}

// Given a slice of bytes, replace any timePiece variables in them with
// the values of the timePiece struct passed in
func ReplaceTime(contents []byte, timePiece timepiece.TimePiece) []byte {
	lumberjack.Info("Replacing time vars in config")

	// using reflection, try to replace any var that shares a field name with
	// the TimePiece struct
	newContents := string(contents)
	piecesOfTime := reflect.ValueOf(&timePiece).Elem()

	for i := 0; i < piecesOfTime.NumField(); i++ {
		fieldName := piecesOfTime.Type().Field(i).Name
		fieldValue := piecesOfTime.Field(i)

		newContents = strings.Replace(
			newContents,
			"$"+fieldName,
			fmt.Sprintf("%v", fieldValue),
			-1)
	}

	return []byte(newContents)
}

// Given a slice of bytes from a JSON config file, unmarshal from JSON into
// a map of interfaces
func UnmarshalConfig(configContents []byte) map[string]interface{} {
	lumberjack.Info("Unmarshalling config from JSON")

	// parse the JSON and return it
	var configJson interface{}
	err := json.Unmarshal(configContents, &configJson)
	if err != nil {
		lumberjack.Fatal("Failed to Unmarshal JSON; contents: %v", configJson)
	}

	return configJson.(map[string]interface{})
}
