package main

import (
	"errors"
	"testing"
)

func TestSerialize(t *testing.T) {
	i := 12
	if Serialize(i) != ":12\r\n" {
		t.Error("Expected other result for Serialize(12)")
	}

	n := -5
	if Serialize(n) != ":-5\r\n" {
		t.Error("Expected other result for Serialize(-5)")
	}

	str := "hello there!"
	if Serialize(str) != "$12\r\nhello there!\r\n" {
		t.Error("Expected other result for Serialize('hello there!')")
	}

	str = ""
	if Serialize(str) != "$0\r\n\r\n" {
		t.Error("Expected other result for Serialize('')")
	}

	arr := []any{3, "word", -1}
	if Serialize(arr) != "*3\r\n:3\r\n$4\r\nword\r\n:-1\r\n" {
		t.Error("Expected other result for Serialize([3, 'word', -1])")
	}

	arr = []any{}
	if Serialize(arr) != "*0\r\n" {
		t.Error("Expected other result for Serialize([])")
	}

	err := errors.New("custom error")
	if Serialize(err) != "-custom error\r\n" {
		t.Error("Expected other result for Serialize(err)")
	}
}
