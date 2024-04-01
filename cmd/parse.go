package main

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

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

func ParseCommands(message string) ([]Command, error) {
	split, err := sanitize(message)
	if err != nil {
		return nil, err
	}

	if len(split) == 0 {
		return nil, errors.New("empty message")
	}

	commands := []Command{}
	i := 0
	for i < len(split) {
		cmd := split[i]
		switch cmd {
		case "version":
			commands = append(commands, Command{Kind: CmdVersion})
		case "ping":
			commands = append(commands, Command{Kind: CmdPing})
		case "keys":
			commands = append(commands, Command{Kind: CmdKeys})
		case "set":
			// Needs another 2 arguments
			if i+2 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdSet, Key: split[i+1], Value: split[i+2]})
			i += 2
		case "get":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdGet, Key: split[i+1]})
			i += 1
		case "del":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdDel, Key: split[i+1]})
			i += 1
		case "list", "ls":
			commands = append(commands, Command{Kind: CmdList})
		case "qadd":
			if i+2 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdQueueAdd, Key: split[i+1], Value: split[i+2]})
			i += 2
		case "qpop":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdQueuePop, Key: split[i+1]})
			i += 1
		case "qlen":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdQueueLen, Key: split[i+1]})
			i += 1
		case "pqadd":
			if i+2 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}

			pqadd := Command{Kind: CmdPQAdd, Key: split[i+1], Value: split[i+2], Priority: 1}
			// TODO: Refactor this
			if i+3 < len(split) {
				if priority, err := strconv.Atoi(split[i+3]); err == nil {
					pqadd.Priority = priority
					i += 3
					commands = append(commands, pqadd)
				} else {
					commands = append(commands, pqadd)
					i += 2
				}

			} else {
				commands = append(commands, pqadd)
				i += 2
			}
		case "pqpop":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdPQPop, Key: split[i+1]})
			i += 1
		case "pqlen":
			if i+1 > len(split) {
				return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
			}
			commands = append(commands, Command{Kind: CmdPQLen, Key: split[i+1]})
			i += 1
		default:
			if cmd == "" {
				break
			}
			return nil, fmt.Errorf("unknown command: '%s'", cmd)
		}
		i++
	}

	return commands, nil
}
