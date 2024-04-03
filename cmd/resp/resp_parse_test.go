package resp

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestParseBulkString(t *testing.T) {
	str := "$5\r\nhello\r\n"
	r := bufio.NewReader(strings.NewReader(str))
	v, err := Read(r)

	if err != nil || v != "hello" {
		t.Error("Expected result to be 'hello'")
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
