package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

var Queues = map[string]*Queue{}
var PQueues = map[string]*PriorityQueue{}

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
	case "pqadd":
		var pq *PriorityQueue
		name := split[1]
		pq, found := PQueues[name]
		if !found {
			pq = NewPriorityQueue()
		}

		priority, err := strconv.Atoi(split[3])
		if err != nil {
			res = fmt.Sprintln("Could not convert", split[3], "to an integer")
			break
		}

		pq.Enqueue(split[2], priority)
		if !found {
			PQueues[name] = pq
		}
	case "pqpop":
		name := split[1]
		pq, found := PQueues[name]
		if !found {
			break
		}

		res = fmt.Sprintln(pq.Dequeue())
	case "pqlen":
		name := split[1]
		pq, found := PQueues[name]
		if !found {
			break
		}

		res = fmt.Sprintln(pq.Length)
	default:
		res = fmt.Sprintf("Unknown command '%s'\n", cmd)
	}

	conn.Write([]byte(res))
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	totalBytes := 0
	messageBytes := []byte{}
	for {
		buffer := make([]byte, 1024)
		// Read data from the client
		n, err := conn.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			panic(err)
		}

		totalBytes += n
		messageBytes = append(messageBytes, buffer[:n]...)
		message := strings.ReplaceAll(string(messageBytes), "\n", "")

		// Only execute when encountering an EOF
		if err != nil {
			execute(message, conn)
			return
		}
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
