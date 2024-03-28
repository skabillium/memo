package main

import (
	"fmt"
	"net"
	"strings"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

var Queues = map[string]*Queue{}

func execute(message string, conn net.Conn) {
	split := strings.Split(message, " ")
	cmd := split[0]

	var res string
	switch cmd {
	case "version":
		res = "Memo server version " + MemoVersion + "\n"
	case "set":
		Set(split[1], split[2])
	case "get":
		value, found := Get(split[1])
		if !found {
			res = "<nil>\n"
		} else {
			res = fmt.Sprintf("%s\n", value)
		}
	case "list", "ls":
		entries := List()
		for i := 0; i < len(entries); i++ {
			res += fmt.Sprintf("'%s':'%s'\n", entries[i][0], entries[i][1])
		}
	case "del":
		Delete(split[1])
	case "qadd":
		var q *Queue
		name := split[1]
		q, found := Queues[name]
		if !found {
			q = NewQueue()
		}

		q.Enqueue(split[2])
		if !found {
			Queues[name] = q
		}
	case "qpop":
		name := split[1]
		q, found := Queues[name]
		if !found {
			break
		}

		res = fmt.Sprintln(q.Dequeue())
	case "qlen":
		name := split[1]
		q, found := Queues[name]
		if !found {
			break
		}

		res = fmt.Sprintln(q.Length)
	default:
		res = fmt.Sprintf("Unknown command '%s'\n", cmd)
	}

	conn.Write([]byte(res))
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Connected to Memo server version " + MemoVersion + "\n"))

	buffer := make([]byte, 1024)
	for {
		// Read data from the client
		n, err := conn.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				return
			}
			panic(err)
		}

		message := strings.ReplaceAll(string(buffer[:n]), "\n", "")
		execute(message, conn)
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:"+DefaultPort)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Memo server is listening on port %s\n", DefaultPort)
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go handleClient(conn)
	}
}
