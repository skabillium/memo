package main

import (
	"reflect"
	"strconv"
)

/*
TODO:
- Find out a way to distinguish simple strings for responses like "OK"
*/
func Serialize(v any) string {
	tp := reflect.TypeOf(v)
	switch tp.Kind() {
	case reflect.Int:
		return SerializeInt(v.(int))
	case reflect.String:
		return SerializeStr(v.(string))
	case reflect.Slice:
		arr := reflect.ValueOf(v)
		out := "*" + strconv.Itoa(arr.Len()) + "\r\n"
		for i := 0; i < arr.Len(); i++ {
			el := arr.Index(i).Interface()
			out += Serialize(el)
		}
		return out
	default:
		if err, ok := v.(error); ok {
			return SerializeError(err)
		}
	}

	return ""
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

func SerializeArr(arr []string) string {
	out := "*"
	if len(arr) == 0 {
		out += "0\r"
		return out
	}

	out += strconv.Itoa(len(arr)) + "\r\n"
	for _, s := range arr {
		out += SerializeSimpleStr(s) + "\n"
	}

	return out
}
