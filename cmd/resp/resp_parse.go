package resp

import (
	"bufio"
	"errors"
	"strconv"
	"strings"
)

// Resp protocol's data types
const (
	RespStatus    = '+' // +<string>\r\n
	RespError     = '-' // -<string>\r\n
	RespString    = '$' // $<length>\r\n<bytes>\r\n
	RespInt       = ':' // :<number>\r\n
	RespNil       = '_' // _\r\n
	RespFloat     = ',' // ,<floating-point-number>\r\n (golang float)
	RespBool      = '#' // true: #t\r\n false: #f\r\n
	RespBlobError = '!' // !<length>\r\n<bytes>\r\n
	RespVerbatim  = '=' // =<length>\r\nFORMAT:<bytes>\r\n
	RespBigInt    = '(' // (<big number>\r\n
	RespArray     = '*' // *<len>\r\n... (same as resp2)
	RespMap       = '%' // %<len>\r\n(key)\r\n(value)\r\n... (golang map)
	RespSet       = '~' // ~<len>\r\n... (same as Array)
	RespAttr      = '|' // |<len>\r\n(key)\r\n(value)\r\n... + command reply
	RespPush      = '>' // ><len>\r\n... (same as Array)
)

func Read(r *bufio.Reader) (any, error) {
	l, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line := strings.Trim(l, "\r\n")

	switch line[0] {
	case RespNil:
		return nil, nil
	case RespBool:
		if line[1] == 't' {
			return true, nil
		}
		return false, nil
	case RespInt:
		return strconv.Atoi(line[1:])
	case RespStatus:
		return line[1:], nil
	case RespString:
		return readString(r, line)
	case RespError:
		return errors.New(line[1:]), nil
	case RespArray:
		return readSlice(r, line)
	}

	return line, nil
}

func readString(r *bufio.Reader, line string) (string, error) {
	n, err := replyLen(line)
	if err != nil {
		return "", err
	}

	b := make([]byte, n+2)
	_, err = r.Read(b)
	if err != nil {
		return "", err
	}

	return string(b[:n]), nil
}

func readSlice(r *bufio.Reader, line string) ([]any, error) {
	n, err := replyLen(line)
	if err != nil {
		return nil, err
	}

	arr := make([]any, n)
	for i := 0; i < len(arr); i++ {
		v, err := Read(r)
		if err != nil {
			return arr, err
		}

		arr[i] = v
	}

	return arr, nil
}

func replyLen(line string) (int, error) {
	n, err := strconv.Atoi(line[1:])
	if err != nil {
		return 0, err
	}

	return n, nil
}
