package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

const DefaultUser = "memo"
const DefaultPassword = "password"
const CurrentRespVersion = "2"

type MemoContext struct {
	conn net.Conn
	rw   *bufio.ReadWriter
}

func NewMemoContext(conn net.Conn) *MemoContext {
	return &MemoContext{
		conn: conn,
		rw:   bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
	}
}

func (c *MemoContext) Writeln(message string) {
	c.rw.WriteString(SerializeSimpleStr(message) + "\n")
}

func (c *MemoContext) Write(message any) {
	c.rw.WriteString(Serialize(message))
}

func (c *MemoContext) Readline() (string, error) {
	return c.rw.ReadString('\n')
}

func (c *MemoContext) Simple(message string) {
	c.rw.WriteString(SerializeSimpleStr(message))
}

func (c *MemoContext) EndWith(message string) {
	c.Writeln(message)
	c.End()
}

func (c *MemoContext) EndWithError(err error) {
	c.Error(err)
	c.End()
}

func (c *MemoContext) End() {
	c.rw.Flush()
}

func (c *MemoContext) Error(err error) {
	c.rw.WriteString(SerializeError(err) + "\n")
}

type MemoEntry struct {
	str string
}

func NewMemoEntry(str string) *MemoEntry {
	return &MemoEntry{str: str}
}

func (e *MemoEntry) String() string {
	return e.str
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

func Execute(ctx *MemoContext, message string) {
	commands, err := ParseCommands(message)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, cmd := range commands {
		switch cmd.Kind {
		case CmdVersion:
			ctx.Writeln(MemoVersion)
		case CmdPing:
			ctx.Writeln("PONG")
		case CmdKeys:
			keys := []string{}
			for k := range KV {
				keys = append(keys, k)
			}
			for q := range Queues {
				keys = append(keys, q)
			}
			for q := range PQueues {
				keys = append(keys, q)
			}

			ctx.Write(keys)
		case CmdSet:
			Set(cmd.Key, cmd.Value)
			ctx.Simple("OK")
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

	// Auth
	user     string
	password string
}

func NewServer(port string) *Server {
	return &Server{
		port:   port,
		quitCh: make(chan struct{}),
	}
}

func (s *Server) Auth(user string, password string) {
	s.user = user
	s.password = password
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

	// Read initialization message
	ctx := NewMemoContext(conn)
	init, err := ctx.Readline()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = s.checkInitLine(init)
	if err != nil {
		ctx.EndWithError(err)
		return
	}

	for {
		line, err := ctx.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Error while reading data for connection", err)
			break
		}

		Execute(ctx, line)
		ctx.End()
	}

}

func (s *Server) checkInitLine(line string) error {
	if s.user == "" || s.password == "" {
		return nil
	}

	split := strings.Split(line, " ")

	if len(split) != 4 {
		return errors.New("ERR wrong number of arguments")
	}

	if split[0] != "hello" {
		return errors.New("ERR invalid command " + split[0])
	}

	if split[1] != CurrentRespVersion {
		return errors.New("NOPROTO unsupported protocol version")
	}

	if s.user != split[2] || s.password != split[3] {
		return errors.New("ERR invalid credentials")
	}

	return nil
}

func main() {
	var (
		port       string
		portSr     string
		user       string
		userSr     string
		password   string
		passwordSr string
	)

	flag.StringVar(&port, "port", "", "Port to run server")
	flag.StringVar(&portSr, "p", "", "Shorthand for port")
	flag.StringVar(&user, "user", "", "User for authentication")
	flag.StringVar(&userSr, "u", "", "Shorthand for user")
	flag.StringVar(&password, "password", "", "Password for authentication")
	flag.StringVar(&passwordSr, "pwd", "", "Shorthand for password")
	flag.Parse()

	if port == "" {
		if portSr != "" {
			port = portSr
		} else {
			port = DefaultPort
		}
	}

	if user == "" {
		if userSr != "" {
			user = userSr
		} else {
			user = DefaultUser
		}
	}

	if password == "" {
		if passwordSr != "" {
			password = passwordSr
		} else {
			password = DefaultPassword
		}
	}

	server := NewServer(DefaultPort)
	server.Auth(user, password)
	server.Start()
}
