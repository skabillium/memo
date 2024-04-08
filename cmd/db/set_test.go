package db

import "testing"

func TestSet(t *testing.T) {
	set := NewSet()

	set.Add("one")
	set.Add("two")
	set.Add("three")

	if set.Size != 3 {
		t.Error("Expected Size to be 3")
	}
	if set.Has("four") {
		t.Error("Expected Has('four') to return false")
	}
	if !set.Has("three") {
		t.Error("Expected Has('three') to return true")
	}

	if set.Delete("four") {
		t.Error("Expected Delete('four') to return false")
	}

	set.Add("one")

	if set.Size != 3 {
		t.Error("Expected Size to be 3")
	}
	if !set.Delete("one") {
		t.Error("Expected Delete('one') to return true")
	}
	if !set.Delete("two") {
		t.Error("Expected Delete('two') to return true")
	}
	if !set.Delete("three") {
		t.Error("Expected Delete('three') to return true")
	}

	if set.Size != 0 {
		t.Error("Expexted Size to be 0")
	}
}
