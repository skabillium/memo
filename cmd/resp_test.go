package main

import (
	"errors"
	"testing"
)

func TestSerialize(t *testing.T) {
	if r, err := Serialize(nil); r != "$-1\r\n" || err != nil {
		t.Error("Expected other result for Serialize(nil)")
	}

	if r, err := Serialize(12); r != ":12\r\n" || err != nil {
		t.Error("Expected other result for Serialize(12)")
	}

	if r, err := Serialize(-5); r != ":-5\r\n" || err != nil {
		t.Error("Expected other result for Serialize(-5)")
	}

	if r, err := Serialize("hello there!"); r != "+hello there!\r\n" || err != nil {
		t.Error("Expected other result for Serialize('hello there!')")
	}

	if r, err := Serialize(""); r != "+\r\n" || err != nil {
		t.Error("Expected other result for Serialize('')")
	}

	arr := []any{3, "word", -1}
	if r, err := Serialize(arr); r != "*3\r\n:3\r\n+word\r\n:-1\r\n" || err != nil {
		t.Error("Expected other result for Serialize([3, 'word', -1])")
	}

	arr = []any{}
	if r, err := Serialize(arr); r != "*0\r\n" || err != nil {
		t.Error("Expected other result for Serialize([])")
	}

	err := errors.New("custom error")
	if r, err := Serialize(err); r != "-custom error\r\n" || err != nil {
		t.Error("Expected other result for Serialize(err)")
	}

	mstr := MemoString("message")
	if r, err := Serialize(mstr); r != "$7\r\nmessage\r\n" || err != nil {
		t.Error("Expected other result for Serialize(MemoString('message'))")
	}

	mstr = MemoString("")
	if r, err := Serialize(mstr); r != "$0\r\n\r\n" || err != nil {
		t.Error("Expected other result for Serialize(MemoString(''))")
	}
}

func TestSerializeStruct(t *testing.T) {}

func TestSerializeMap(t *testing.T) {}
