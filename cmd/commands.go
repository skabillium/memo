package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func ErrUnknownCmd(cmd string) error {
	return fmt.Errorf("ERR unknown command '%s'", cmd)
}

func ErrInvalidNArg(cmd string) error {
	return fmt.Errorf("invalid number of arguments for command '%s'", cmd)
}

var ErrNotInt = errors.New("ERR value is not an integer or out of range")
var ErrUnbalancedQuotes = errors.New("ERR unbalanced quotes")

type CommandType = byte

const (
	// Server commands
	CmdVersion CommandType = iota
	CmdPing
	CmdKeys
	CmdAuth
	CmdHello
	CmdInfo
	CmdDbSize
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
	// Sets
	CmdSetAdd
	CmdSetMembers
	CmdSetRem
	CmdSetIsMember
	CmdSetInter
	CmdSetCard
)

type AuthOptions struct {
	User     string
	Password string
}

type Command struct {
	Kind   CommandType
	Key    string
	Keys   []string
	Value  string
	Values []string

	Pattern     string      // keys
	ExpireIn    int         // expire
	Priority    int         // pqadd
	Auth        AuthOptions // hello
	RespVersion string      // hello
	Limit       int
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
		if argc > 2 {
			return nil, ErrInvalidNArg(cmd)
		}

		keys := &Command{Kind: CmdKeys, Pattern: "*"}
		if argc == 2 {
			keys.Pattern = split[1]
		}
		return keys, nil
	case "info":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdInfo}, nil
	case "dbsize":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdDbSize}, nil
	case "flushall":
		if argc != 1 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdFlushAll}, nil
	case "cleanup":
		cleanup := &Command{Kind: CmdCleanup}
		if argc == 2 {
			limit, err := strconv.Atoi(split[1])
			if err != nil {
				return nil, ErrNotInt
			}

			cleanup.Limit = limit
		}

		return cleanup, nil
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
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		set := &Command{Kind: CmdSet, Key: split[1], Value: split[2]}
		if argc > 3 && strings.ToLower(split[3]) == "ex" {
			// Parse expiration option
			if argc != 5 {
				return nil, errors.New("invalid number of arguments for expiration option")
			}
			seconds, err := strconv.Atoi(split[4])
			if err != nil {
				return nil, ErrNotInt
			}
			set.ExpireIn = seconds
		}
		return set, nil
	case "get":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdGet, Key: split[1]}, nil
	case "del":
		if argc < 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdDel, Keys: split[1:]}, nil
	case "qadd":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		qadd := &Command{Kind: CmdQueueAdd, Key: split[1], Priority: 1}

		for i := 2; i < argc; i++ {
			if i+1 < argc && strings.ToLower(split[i]) == "pr" {
				priority, err := strconv.Atoi(split[i+1])
				if err != nil {
					return nil, ErrNotInt
				}

				qadd.Priority = priority
				i++
				continue
			}
			qadd.Values = append(qadd.Values, split[i])
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
	case "sadd":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetAdd, Key: split[1], Values: split[2:]}, nil
	case "smembers":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetMembers, Key: split[1]}, nil
	case "srem":
		if argc < 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetRem, Key: split[1], Values: split[2:]}, nil
	case "sismember":
		if argc != 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetIsMember, Key: split[1], Value: split[2]}, nil
	case "scard":
		if argc != 2 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetCard, Key: split[1]}, nil
	case "sinter":
		if argc != 3 {
			return nil, ErrInvalidNArg(cmd)
		}
		return &Command{Kind: CmdSetInter, Keys: []string{split[1], split[2]}}, nil
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
				return nil, ErrUnbalancedQuotes
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
