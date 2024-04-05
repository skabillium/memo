package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type CommandType = byte

func ErrUnknownCmd(cmd string) error {
	return fmt.Errorf("ERR unknown command '%s'", cmd)
}

func ErrInvalidNArg(cmd string) error {
	return fmt.Errorf("invalid number of arguments for command '%s'", cmd)
}

const (
	// Server commands
	CmdVersion CommandType = iota
	CmdPing
	CmdKeys
	CmdAuth
	CmdHello
	CmdInfo
	CmdFlushAll
	CmdCleanup
	CmdExpire
	// KV
	CmdSet
	CmdGet
	CmdList
	CmdDel
	// Priority Queues
	CmdQueueAdd
	CmdQueuePop
	CmdQueueLen
	// Lists
	CmdLPush
	CmdLPop
	CmdRPush
	CmdRPop
	CmdLLen
)

type AuthOptions struct {
	User     string
	Password string
}

type Command struct {
	Kind   CommandType
	Key    string
	Value  string
	Values []string

	ExpireIn    int         // expire
	Priority    int         // pqadd
	Auth        AuthOptions // hello
	RespVersion string      // hello
}

func ParseCommand(message string) (*Command, error) {
	split, err := sanitize(message)
	if err != nil {
		return nil, err
	}

	argc := len(split)
	if argc == 0 {
		return nil, errors.New("empty message")
	}

	cmd := strings.ToLower(split[0])
	switch cmd {
	case "version":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdVersion}, nil
	case "ping":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdPing}, nil
	case "keys":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdKeys}, nil
	case "info":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdInfo}, nil
	case "flushall":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdFlushAll}, nil
	case "cleanup":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdCleanup}, nil
	case "expire":
		if argc != 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		seconds, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}
		return &Command{Kind: CmdExpire, Key: split[1], ExpireIn: seconds}, nil
	case "auth":
		if argc != 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdAuth, Auth: AuthOptions{User: split[1], Password: split[2]}}, nil
	case "hello":
		if argc < 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		hello := &Command{Kind: CmdHello, RespVersion: split[1]}
		if argc > 2 && strings.ToLower(split[2]) == "auth" {
			if argc != 5 {
				return nil, ErrInvalidNArg(cmd)
			}

			hello.Auth.User = split[3]
			hello.Auth.Password = split[4]
		}
		return hello, nil
	case "set":
		if argc != 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSet, Key: split[1], Value: split[2]}, nil
	case "get":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdGet, Key: split[1]}, nil
	case "del":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdDel, Key: split[1]}, nil
	case "qadd":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		qadd := &Command{Kind: CmdQueueAdd, Key: split[1], Value: split[2], Priority: 1}
		if argc == 4 {
			priority, err := strconv.Atoi(split[3])
			if err != nil {
				return nil, err
			}
			qadd.Priority = priority
		}
		return qadd, nil
	case "qpop":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdQueuePop, Key: split[1]}, nil
	case "qlen":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdQueueLen, Key: split[1]}, nil
	case "lpush":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		lpush := &Command{Kind: CmdLPush, Key: split[1], Values: []string{}}
		lpush.Values = append(lpush.Values, split[2:]...)
		return lpush, nil
	case "lpop":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdLPop, Key: split[1]}, nil
	case "rpush":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		rpush := &Command{Kind: CmdRPush, Key: split[1], Values: []string{}}
		rpush.Values = append(rpush.Values, split[2:]...)
		return rpush, nil
	case "rpop":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdRPop, Key: split[1]}, nil
	case "llen":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdLLen, Key: split[1]}, nil
	}

	return nil, ErrUnknownCmd(cmd)
}

func isWhitespace(b byte) bool {
	return unicode.IsSpace(rune(b))
}

func sanitize(message string) ([]string, error) {
	out := []string{}
	i := 0

	for i < len(message) {
		c := message[i]
		if isWhitespace(c) {
			i++
			continue
		}

		if c == '"' || c == '\'' {
			term := c
			i++
			start := i
			c := message[i]
			for i < len(message) && c != term {
				// TODO: Parse special characters
				c = message[i]
				i++
			}

			if c != term {
				return nil, errors.New("unterminated string")
			}

			out = append(out, message[start:i-1])
			i++
			continue
		}

		// TODO: Refactor this
		start := i
		for i < len(message) && !isWhitespace(c) {
			c = message[i]
			i++
		}

		if i == len(message) && !isWhitespace(c) {
			i++
		}

		out = append(out, message[start:i-1])
	}

	return out, nil
}
