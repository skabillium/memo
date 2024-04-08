// Provide serialization functions for compliance with the REdis Serialization Protocol
// specification, see: https://redis.io/docs/reference/protocol-spec/#resp-protocol-description
package resp

import (
	"fmt"
	"reflect"
	"strconv"
)

type SimpleString string

func Serialize(v any) (string, error) {
	if v == nil {
		return SerializeNil(), nil
	}

	tp := reflect.TypeOf(v)
	switch tp.Kind() {
	case reflect.Bool:
		return SerializeBool(v.(bool)), nil
	case reflect.Int:
		return SerializeInt(v.(int)), nil
	case reflect.String:
		if tp == reflect.TypeOf(SimpleString("")) {
			return SerializeSimpleStr(string(v.(SimpleString))), nil
		}
		return SerializeStr(v.(string)), nil
	case reflect.Struct:
		stc := reflect.ValueOf(v)
		out := "%" + fmt.Sprint(stc.NumField()) + "\r\n"
		for i := 0; i < stc.NumField(); i++ {
			fieldValue := stc.Field(i)
			fieldName := tp.Field(i).Name

			out += SerializeStr(fieldName)
			r, err := Serialize(fieldValue.Interface())
			if err != nil {
				return "", err
			}
			out += r
		}
		return out, nil
	case reflect.Map:
		mp := reflect.ValueOf(v)
		keys := mp.MapKeys()

		out := "%" + fmt.Sprint(mp.Len()) + "\r\n"
		for _, k := range keys {
			val := mp.MapIndex(k)
			r, err := Serialize(k.Interface())
			if err != nil {
				return "", err
			}
			out += r

			r, err = Serialize(val.Interface())
			if err != nil {
				return "", err
			}
			out += r
		}
		return out, nil
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

	// TODO: Maybe skip if it cannot be serialized
	return "", fmt.Errorf("value '%s' cannot be serialized", tp)
}

func SerializeNil() string {
	return "$-1\r\n"
}

func SerializeBool(b bool) string {
	out := "#"
	if b {
		out += "t"
	} else {
		out += "f"
	}
	out += "\r\n"
	return out
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
