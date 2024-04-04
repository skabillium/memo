package main

import "testing"

func TestList(t *testing.T) {
	list := NewList()
	list.Append("2")
	list.Append("5")
	list.Append("7")

	if list.Length != 3 {
		t.Error("Expected Length to be 3")
	}
	if list.PeekTail() != "7" {
		t.Error("Expected PeekTail() to return 7")
	}
	if list.PeekHead() != "2" {
		t.Error("Expexted PeekHead() to return 2")
	}

	list.Prepend("1")
	if list.Length != 4 {
		t.Error("Expected Length to be 4")
	}
	if list.PeekHead() != "1" {
		t.Error("Expected PeekHead() to return 1")
	}

	if list.PopTail() != "7" {
		t.Error("Expexted PopTail() to return 7")
	}
	if list.PopTail() != "5" {
		t.Error("Expexted PopTail() to return 5")
	}
	if list.PopTail() != "2" {
		t.Error("Expexted PopTail() to return 2")
	}
	if list.Length != 1 {
		t.Error("Expected Length to be 1")
	}
	if list.PopTail() != "1" {
		t.Error("Expected PopTail() to return 1")
	}
	if list.Length != 0 {
		t.Error("Expected Length to be 0")
	}

	list.Append("12")
	list.Append("13")
	if list.Length != 2 {
		t.Error("Expected Length to be 2")
	}
	if list.PopHead() != "12" {
		t.Error("Expected PopHead() to return 12")
	}
	if list.PopHead() != "13" {
		t.Error("Expected PopHead() to return 13")
	}
	if list.Length != 0 {
		t.Error("Expected Length to be 0")
	}

	if list.PeekHead() != "" {
		t.Error("Expected PeekHead() to return ''")
	}
	if list.PeekTail() != "" {
		t.Error("Expected PeekTail() to return ''")
	}
}
