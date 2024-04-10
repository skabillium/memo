package main

import (
	"reflect"
	"testing"
)

func TestSplitTokens(t *testing.T) {
	if res, err := splitTokens("version"); err != nil || !reflect.DeepEqual(res, []string{"version"}) {
		t.Error("Expected sanitize('version') to return ['version']")
	}
	if res, err := splitTokens("version\n"); err != nil || !reflect.DeepEqual(res, []string{"version"}) {
		t.Error("Expected sanitize('version') to return ['version']")
	}
	if res, err := splitTokens("set name bill\nset age 22"); err != nil || !reflect.DeepEqual(res, []string{
		"set", "name", "bill", "set", "age", "22",
	}) {
		t.Error("Expected other result for multiple operations")
	}

	if res, err := splitTokens("set \"my message\" \"Hello there!\""); err != nil || !reflect.DeepEqual(res, []string{
		"set", "my message", "Hello there!",
	}) {
		t.Error("Expected other result for string input")
	}

	_, err := splitTokens("set \"error")
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

func TestParseServerCommands(t *testing.T) {
	var str string
	var cmd, res *Command

	str = "version"
	cmd = &Command{Kind: CmdVersion}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "ping"
	cmd = &Command{Kind: CmdPing}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "keys"
	cmd = &Command{Kind: CmdKeys, Pattern: "*"}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "keys user:*"
	cmd = &Command{Kind: CmdKeys, Pattern: "user:*"}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "info"
	cmd = &Command{Kind: CmdInfo}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
}

func TestParseHelloCommand(t *testing.T) {
	var str string
	var cmd, res *Command

	if _, err := ParseCommand("hello"); err == nil {
		t.Error("Expected 'hello' to return parsing error")
	}

	str = "hello 3"
	cmd = &Command{Kind: CmdHello, RespVersion: "3"}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "hello 3 auth usr pwd"
	cmd = &Command{
		Kind:        CmdHello,
		RespVersion: "3",
		Auth: AuthOptions{
			User:     "usr",
			Password: "pwd",
		},
	}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	if _, err := ParseCommand("hello 3 auth user"); err == nil {
		t.Error("Expected 'hello 3 auth user' to return parsing error")
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

func TestParseListCommands(t *testing.T) {
	var str string
	var cmd, res *Command

	str = "lpush list 1 2 3"
	cmd = &Command{Kind: CmdLPush, Key: "list", Values: []string{"1", "2", "3"}}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "rpush list 1 2 3"
	cmd = &Command{Kind: CmdRPush, Key: "list", Values: []string{"1", "2", "3"}}
	res, _ = ParseCommand(str)
	if !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "lpop list"
	cmd = &Command{Kind: CmdLPop, Key: "list"}
	if res, _ := ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "rpop list"
	cmd = &Command{Kind: CmdRPop, Key: "list"}
	if res, _ := ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "llen list"
	cmd = &Command{Kind: CmdLLen, Key: "list"}
	if res, _ := ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
}

func TestParseSetCommands(t *testing.T) {
	var str string
	var cmd, res *Command

	str = "sadd sett 1 2 3"
	cmd = &Command{Kind: CmdSetAdd, Key: "sett", Values: []string{"1", "2", "3"}}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "srem sett 1"
	cmd = &Command{Kind: CmdSetRem, Key: "sett", Values: []string{"1"}}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "sismember sett 1"
	cmd = &Command{Kind: CmdSetIsMember, Key: "sett", Value: "1"}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "scard sett"
	cmd = &Command{Kind: CmdSetCard, Key: "sett"}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}

	str = "sinter set1 set2"
	cmd = &Command{Kind: CmdSetInter, Keys: []string{"set1", "set2"}}
	if res, _ = ParseCommand(str); !reflect.DeepEqual(cmd, res) {
		t.Error("Expected result to be:", cmd, "got", res)
	}
}
