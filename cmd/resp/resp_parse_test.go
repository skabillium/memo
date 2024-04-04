package resp

import (
	"bufio"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseNil(t *testing.T) {
	str := "_\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)
	if err != nil || v != nil {
		t.Error("Expected result to be nil")
	}
}

func TestParseBool(t *testing.T) {
	str := "#t\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)
	if err != nil || v != true {
		t.Error("Expected result to be true")
	}

	str = "#f\r\n"
	r = bufio.NewReader(strings.NewReader(str))
	v, err = Read(r)
	if err != nil || v != false {
		t.Error("Expected result to be true")
	}
}

func TestParseInt(t *testing.T) {
	str := ":15\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)
	if err != nil || v != 15 {
		t.Error("Expected result to be 15")
	}

	str = ":0\r\n"
	r = bufio.NewReader(strings.NewReader(str))
	v, err = Read(r)
	if err != nil || v != 0 {
		t.Error("Expected result to be 0")
	}

	str = ":-2\r\n"
	r = bufio.NewReader(strings.NewReader(str))
	v, err = Read(r)
	if err != nil || v != -2 {
		t.Error("Expected result to be -2")
	}
}

func TestParseBulkString(t *testing.T) {
	str := "$5\r\nhello\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)
	if err != nil || v != "hello" {
		t.Error("Expected result to be 'hello'")
	}
}

func TestParseError(t *testing.T) {
	str := "-ERR\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)
	if err != nil || !reflect.DeepEqual(v, errors.New("ERR")) {
		t.Error("Expected result to be error('ERR')")
	}
}

func TestParseArray(t *testing.T) {
	str := "*2\r\n$2\r\nhi\r\n$2\r\nlo\r\n"
	r := bufio.NewReader(strings.NewReader(str))

	expected := []any{"hi", "lo"}
	v, err := Read(r)

	if err != nil || !reflect.DeepEqual(v, expected) {
		t.Error("Expected other result for array")
	}
}
