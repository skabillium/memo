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
	case "ping":
		res = "pong\n"
	case "keys":
		for k := range KV {
			res += k + "\n"
		}
		for q := range Queues {
			res += q + "\n"
		}
		for q := range PQueues {
			res += q + "\n"
		}
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

type Server struct {
	port   string
	ln     net.Listener
	quitCh chan struct{}
}

func NewServer(port string) *Server {
	return &Server{port: port, quitCh: make(chan struct{})}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", "localhost:"+s.port)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Println("Memo server started on port", s.port)

	s.ln = ln
	go s.acceptLoop()
	<-s.quitCh

	return nil
}

// TODO: Maybe keep connection open instead of request-response pattern
func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	totalBytes := 0
	payload := []byte{}
	for {
		buffer := make([]byte, 1024)
		// Read data from the client
		n, err := conn.Read(buffer)
		if err != nil && err.Error() != "EOF" {
			panic(err)
		}

		totalBytes += n
		payload = append(payload, buffer[:n]...)
		message := strings.ReplaceAll(string(payload), "\n", "")

		// Only execute when encountering an EOF
		if err != nil {
			execute(message, conn)
			return
		}
	}

}

func main() {
	server := NewServer(DefaultPort)
	server.Start()
}
