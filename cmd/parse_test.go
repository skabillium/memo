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
	message := `
	ping
	version
	set name bill
	qadd emails em@example.com
	qadd emails other@example.com 2
	`
	expected := []Command{
		{Kind: CmdPing},
		{Kind: CmdVersion},
		{Kind: CmdSet, Key: "name", Value: "bill"},
		{Kind: CmdQueueAdd, Key: "emails", Value: "em@example.com", Priority: 1},
		{Kind: CmdQueueAdd, Key: "emails", Value: "other@example.com", Priority: 2},
	}

	commands, err := ParseCommands(message)
	if err != nil || !reflect.DeepEqual(expected, commands) {
		t.Error("Expected Parse commands to return", expected, "got", commands)
	}
}
