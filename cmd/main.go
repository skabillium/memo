package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

type MemoContext struct {
	conn    net.Conn
	Message string
}

func NewMemoContext(conn net.Conn, message string) *MemoContext {
	return &MemoContext{conn: conn, Message: message}
}

func (c *MemoContext) Writeln(message string) {
	c.conn.Write([]byte(message + " \n"))
}

func (c *MemoContext) Error(err string) {
	c.Writeln("[ERROR]: " + err)
}

var Queues = map[string]*Queue{}
var PQueues = map[string]*PriorityQueue{}

type CommandType = byte

const (
	// Server commands
	CmdVersion CommandType = iota
	CmdPing
	CmdKeys
	// KV
	CmdSet
	CmdGet
	CmdList
	CmdDel
	// Queues
	CmdQueueAdd
	CmdQueuePop
	CmdQueueLen
	// Priority Queues
	CmdPQAdd
	CmdPQPop
	CmdPQLen
)

type Command struct {
	Kind     CommandType
	Key      string
	Value    string
	Priority int
}

func Execute(ctx *MemoContext) {
	commands, err := ParseCommands(ctx.Message)
	if err != nil {
		ctx.Error(err.Error())
		return
	}

	for _, cmd := range commands {
		switch cmd.Kind {
		case CmdVersion:
			ctx.Writeln(MemoVersion)
		case CmdPing:
			ctx.Writeln("pong")
		case CmdKeys:
			var res string
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
		case CmdSet:
			Set(cmd.Key, cmd.Value)
		case CmdGet:
			value, found := Get(cmd.Key)
			if !found {
				ctx.Writeln("<nil>")
				break
			}

			ctx.Writeln(value)
		case CmdList:
			entries := List()
			var res string
			for i := 0; i < len(entries); i++ {
				res += fmt.Sprintf("'%s':'%s'\n", entries[i][0], entries[i][1])
			}
			ctx.Writeln(res)
		case CmdDel:
			Delete(cmd.Key)
		case CmdQueueAdd:
			var q *Queue
			q, found := Queues[cmd.Key]
			if !found {
				q = NewQueue()
			}

			q.Enqueue(cmd.Value)
			if !found {
				Queues[cmd.Key] = q
			}
		case CmdQueuePop:
			q, found := Queues[cmd.Key]
			if !found {
				break
			}

			ctx.Writeln(q.Dequeue())
		case CmdQueueLen:
			q, found := Queues[cmd.Key]
			if !found {
				break
			}

			ctx.Writeln(strconv.Itoa(q.Length))
		case CmdPQAdd:
			var pq *PriorityQueue
			pq, found := PQueues[cmd.Key]
			if !found {
				pq = NewPriorityQueue()
			}

			pq.Enqueue(cmd.Value, cmd.Priority)
			if !found {
				PQueues[cmd.Key] = pq
			}
		case CmdPQPop:
			pq, found := PQueues[cmd.Key]
			if !found {
				break
			}

			ctx.Writeln(pq.Dequeue())
		case CmdPQLen:
			pq, found := PQueues[cmd.Key]
			if !found {
				break
			}

			ctx.Writeln(strconv.Itoa(pq.Length))
		}
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
		if err != nil && err != io.EOF {
			fmt.Println("Error while reading data for connection")
			return
		}

		totalBytes += n
		payload = append(payload, buffer[:n]...)

		// Only execute when encountering an EOF
		if err == io.EOF {
			Execute(NewMemoContext(conn, string(payload)))
			return
		}
	}

}

func main() {
	server := NewServer(DefaultPort)
	server.Start()
}
