package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

type MemoContext struct {
	conn net.Conn
}

func NewMemoContext(conn net.Conn) *MemoContext {
	return &MemoContext{conn: conn}
}

func (c *MemoContext) Writeln(message string) {
	c.conn.Write([]byte(message + "\n"))
}

func (c *MemoContext) Error(err string) {
	c.Writeln("[ERROR]: " + err)
}

var Queues = map[string]*Queue{}
var PQueues = map[string]*PriorityQueue{}

func execute(ctx *MemoContext, message string) {
	split := strings.Split(message, " ")
	cmd := split[0]

	var res string
	switch cmd {
	case "version":
		ctx.Writeln("Memo server version " + MemoVersion)
	case "ping":
		ctx.Writeln("pong")
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
		ctx.Writeln(res)
	case "set":
		if len(split) != 3 {
			ctx.Error("Invalid number of arguments for 'set' command")
			return
		}

		Set(split[1], split[2])
	case "get":
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'get' command")
			return
		}

		value, found := Get(split[1])
		if !found {
			ctx.Writeln("<nil>")
		} else {
			ctx.Writeln(fmt.Sprint(value))
		}
	case "list", "ls":
		entries := List()
		for i := 0; i < len(entries); i++ {
			res += fmt.Sprintf("'%s':'%s'\n", entries[i][0], entries[i][1])
		}
		ctx.Writeln(res)
	case "del":
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'del' command")
			return
		}
		Delete(split[1])
	case "qadd":
		if len(split) != 3 {
			ctx.Error("Incalid number of arguments for 'qadd' command")
			return
		}

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
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'qpop' command")
			return
		}

		name := split[1]
		q, found := Queues[name]
		if !found {
			break
		}

		ctx.Writeln(q.Dequeue())
	case "qlen":
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'qlen' command")
			return
		}

		name := split[1]
		q, found := Queues[name]
		if !found {
			break
		}

		ctx.Writeln(fmt.Sprint(q.Length))
	case "pqadd":
		if len(split) != 4 {
			ctx.Error("Invalid number of arguments for 'pqadd' command")
			return
		}

		var pq *PriorityQueue
		name := split[1]
		pq, found := PQueues[name]
		if !found {
			pq = NewPriorityQueue()
		}

		priority, err := strconv.Atoi(split[3])
		if err != nil {
			ctx.Error(fmt.Sprintf("Could not convert '%s' to an integer", split[3]))
			return
		}

		pq.Enqueue(split[2], priority)
		if !found {
			PQueues[name] = pq
		}
	case "pqpop":
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'pqpop' command")
			return
		}

		name := split[1]
		pq, found := PQueues[name]
		if !found {
			break
		}

		ctx.Writeln(pq.Dequeue())
	case "pqlen":
		if len(split) != 2 {
			ctx.Error("Invalid number of arguments for 'pqlen' command")
			return
		}

		name := split[1]
		pq, found := PQueues[name]
		if !found {
			break
		}

		ctx.Writeln(fmt.Sprint(pq.Length))
	default:
		ctx.Writeln(fmt.Sprintf("Unknown command '%s'", cmd))
	}
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
			fmt.Println("Error while reading data for connection")
			return
		}

		totalBytes += n
		payload = append(payload, buffer[:n]...)
		message := strings.ReplaceAll(string(payload), "\n", "")

		// Only execute when encountering an EOF
		if err != nil {
			execute(NewMemoContext(conn), message)
			return
		}
	}

}

func main() {
	server := NewServer(DefaultPort)
	server.Start()
}
