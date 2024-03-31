package main

import (
	"errors"
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

// TODO: Parse multiple commands
func ParseCommand(message string) (*Command, error) {
	if message == "" {
		return nil, errors.New("empty message")
	}

	split := strings.Split(message, " ")
	cmd := split[0]
	switch cmd {
	case "version":
		return &Command{Kind: CmdVersion}, nil
	case "ping":
		return &Command{Kind: CmdPing}, nil
	case "keys":
		return &Command{Kind: CmdKeys}, nil
	case "set":
		if len(split) != 3 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdSet, Key: split[1], Value: split[2]}, nil
	case "get":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdGet, Key: split[1]}, nil
	case "del":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdDel, Key: split[1]}, nil
	case "list", "ls":
		return &Command{Kind: CmdList}, nil
	case "qadd":
		if len(split) != 3 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdQueueAdd, Key: split[2]}, nil
	case "qpop":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdQueuePop, Key: split[1]}, nil
	case "qlen":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdQueueLen, Key: split[1]}, nil
	case "pqadd":
		if len(split) < 3 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}

		pqadd := &Command{Kind: CmdPQAdd, Key: split[1], Value: split[2], Priority: 1}
		if len(split) == 4 {
			priority, err := strconv.Atoi(split[3])
			if err != nil {
				return nil, fmt.Errorf("could not convert '%s' to integer", split[3])
			}

			if priority <= 0 {
				return nil, fmt.Errorf("received invalid negative value '%d' for priority", priority)
			}

			pqadd.Priority = priority
		}

		return pqadd, nil
	case "pqpop":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdPQPop, Key: split[1]}, nil
	case "pqlen":
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid number of arguments for '%s' command", cmd)
		}
		return &Command{Kind: CmdPQLen, Key: split[1]}, nil
	}

	return nil, fmt.Errorf("unknown command '%s'", cmd)
}

func Execute(ctx *MemoContext, message string) {
	cmd, err := ParseCommand(message)
	if err != nil {
		ctx.Error(err.Error())
		return
	}

	var res string
	switch cmd.Kind {
	case CmdVersion:
		ctx.Writeln(MemoVersion)
	case CmdPing:
		ctx.Writeln("pong")
	case CmdKeys:
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
			return
		}

		ctx.Writeln(fmt.Sprint(value))
	case CmdList:
		entries := List()
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

		ctx.Writeln(fmt.Sprint(q.Length))
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
			Execute(NewMemoContext(conn), message)
			return
		}
	}

}

func main() {
	server := NewServer(DefaultPort)
	server.Start()
}
