package main

import "testing"

func TestStringifyRequest(t *testing.T) {
	req := []any{"set", "message", "hello world!"}
	if str, err := StringifyRequest(req); str != "set message \"hello world!\"" || err != nil {
		t.Error("Expected other result for StringifyRequest")
	}
}
