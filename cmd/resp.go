// Provide serialization functions for compliance with the REdis Serialization Protocol
// specification, see: https://redis.io/docs/reference/protocol-spec/#resp-protocol-description

package main

import (
	"fmt"
	"reflect"
	"strconv"
)

func Serialize(v any) (string, error) {
	tp := reflect.TypeOf(v)
	switch tp.Kind() {
	case reflect.Int:
		return SerializeInt(v.(int)), nil
	case reflect.String:
		if tp == reflect.TypeOf(MemoString("")) {
			return SerializeStr(fmt.Sprint(v)), nil
		}

		return SerializeSimpleStr(v.(string)), nil
	case reflect.Slice:
		arr := reflect.ValueOf(v)
		out := "*" + strconv.Itoa(arr.Len()) + "\r\n"
		for i := 0; i < arr.Len(); i++ {
			el := arr.Index(i).Interface()
			r, err := Serialize(el)
			if err != nil {
				return "", err
			}

			out += r
		}
		return out, nil
	default:
		if err, ok := v.(error); ok {
			return "-" + err.Error() + "\r\n", nil
		}
	}

	return "", fmt.Errorf("value '%s' cannot be serialized", tp)
}

func SerializeSimpleStr(str string) string {
	return "+" + str + "\r\n"
}

func SerializeStr(str string) string {
	out := "$"
	out += strconv.Itoa(len(str)) + "\r\n"
	out += str + "\r\n"
	return out
}

func SerializeError(err error) string {
	return "-" + err.Error() + "\r\n"
}

func SerializeInt(n int) string {
	return ":" + strconv.Itoa(n) + "\r\n"
}
