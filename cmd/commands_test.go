package main

import (
	"reflect"
	"testing"
)

func TestSanitize(t *testing.T) {
	if res, err := sanitize("version"); err != nil || !reflect.DeepEqual(res, []string{"version"}) {
		t.Error("Expected sanitize('version') to return ['version']")
	}
	if res, err := sanitize("version\n"); err != nil || !reflect.DeepEqual(res, []string{"version"}) {
		t.Error("Expected sanitize('version') to return ['version']")
	}
	if res, err := sanitize("set name bill\nset age 22"); err != nil || !reflect.DeepEqual(res, []string{
		"set", "name", "bill", "set", "age", "22",
	}) {
		t.Error("Expected other result for multiple operations")
	}

	if res, err := sanitize("set \"my message\" \"Hello there!\""); err != nil || !reflect.DeepEqual(res, []string{
		"set", "my message", "Hello there!",
	}) {
		t.Error("Expected other result for string input")
	}

	_, err := sanitize("set \"error")
	if err == nil {
		t.Error("Expected unterminated string error")
	}
}

func TestParse(t *testing.T) {
	str := "set name bill"
	cmd := &Command{Kind: CmdSet, Key: "name", Value: "bill"}
	res, _ := ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
}

func TestParseQueueCommands(t *testing.T) {
	var str string
	var cmd, res *Command

	str = "qadd queue 1"
	cmd = &Command{Kind: CmdQueueAdd, Key: "queue", Values: []string{"1"}, Priority: 1}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
	str = "qadd queue 1 pr 2"
	cmd = &Command{Kind: CmdQueueAdd, Key: "queue", Values: []string{"1"}, Priority: 2}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
	str = "qadd queue 1 2 3"
	cmd = &Command{Kind: CmdQueueAdd, Key: "queue", Values: []string{"1", "2", "3"}, Priority: 1}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
	str = "qadd queue 1 2 3 pr 2"
	cmd = &Command{Kind: CmdQueueAdd, Key: "queue", Values: []string{"1", "2", "3"}, Priority: 2}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
}
