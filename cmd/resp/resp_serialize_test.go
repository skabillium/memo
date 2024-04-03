package resp

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

	if r, err := Serialize("hello there!"); r != "$12\r\nhello there!\r\n" || err != nil {
		t.Error("Expected other result for Serialize('hello there!')")
	}

	if r, err := Serialize(""); r != "$0\r\n\r\n" || err != nil {
		t.Error("Expected other result for Serialize('')")
	}

	arr := []any{3, "word", -1}
	if r, err := Serialize(arr); r != "*3\r\n:3\r\n$4\r\nword\r\n:-1\r\n" || err != nil {
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
}

func TestSerializeSimple(t *testing.T) {
	if SerializeSimpleStr("OK") != "+OK\r\n" {
		t.Error("Expected other result from SerializeSimpleStr('OK')")
	}

	if SerializeSimpleStr("") != "+\r\n" {
		t.Error("Expected other result from SerializeSimpleStr('')")
	}
}

type Person struct {
	Name string
	Age  int
}

func TestSerializeStruct(t *testing.T) {
	b := Person{Name: "Bill", Age: 22}
	expected := "%2\r\n$4\r\nName\r\n$4\r\nBill\r\n$3\r\nAge\r\n:22\r\n"
	if r, err := Serialize(b); r != expected || err != nil {
		t.Errorf("Expected '%s' and got '%s'", expected, r)
	}
}
