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

}
