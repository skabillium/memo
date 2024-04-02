package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"strconv"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

const DefaultUser = "memo"
const DefaultPassword = "password"
const CurrentRespVersion = "2"

type MemoContext struct {
	conn    net.Conn
	rw      *bufio.ReadWriter
	hasAuth bool
}

func NewMemoContext(conn net.Conn) *MemoContext {
	return &MemoContext{
		conn:    conn,
		rw:      bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)),
		hasAuth: false,
	}
}

func (c *MemoContext) Authenticate() {
	c.hasAuth = true
}

func (c *MemoContext) Write(message any) {
	payload, err := Serialize(message)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.rw.WriteString(payload)
}

func (c *MemoContext) Readline() (string, error) {
	return c.rw.ReadString('\n')
}

func (c *MemoContext) EndWith(message any) {
	c.Write(message)
	c.End()
}

func (c *MemoContext) End() {
	c.rw.Flush()
}

func (c *MemoContext) Error(err error) {
	c.Write(err)
}

type MemoString string

var Queues = map[string]*Queue{}
var PQueues = map[string]*PriorityQueue{}

func (s *Server) Execute(ctx *MemoContext, message string) {
	commands, err := ParseCommands(message)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !ctx.hasAuth && s.requireAuth {
		if len(commands) == 0 {
			return
		}

		if commands[0].Kind != CmdHello {
			ctx.Error(errors.New("authentication required"))
			return
		}

		if !(commands[0].Auth.User == s.user && commands[0].Auth.Password == s.password) {
			ctx.Error(errors.New("WRONGPASS invalid username-password pair or user is disabled"))
			return
		}

		ctx.Authenticate()
	}

	for _, cmd := range commands {
		switch cmd.Kind {
		case CmdVersion:
			ctx.Write(MemoVersion)
		case CmdPing:
			ctx.Write("PONG")
		case CmdHello:
			ctx.Write("HELLO")
		case CmdKeys:
			keys := []MemoString{}
			for k := range KV {
				keys = append(keys, MemoString(k))
			}
			for q := range Queues {
				keys = append(keys, MemoString(q))
			}
			for q := range PQueues {
				keys = append(keys, MemoString(q))
			}

			ctx.Write(keys)
		case CmdSet:
			Set(cmd.Key, cmd.Value)
			ctx.Write("OK")
		case CmdGet:
			value, found := Get(cmd.Key)
			if !found {
				ctx.Write(nil)
				break
			}

			ctx.Write(value)
		case CmdList:
			ctx.Error(errors.New("ERR unsupported command 'list'"))
			// entries := List()
			// var res string
			// for i := 0; i < len(entries); i++ {
			// 	res += fmt.Sprintf("'%s':'%s'\n", entries[i][0], entries[i][1])
			// }
			// ctx.Writeln(res)
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

			ctx.Write(q.Dequeue())
		case CmdQueueLen:
			q, found := Queues[cmd.Key]
			if !found {
				break
			}

			ctx.Write(strconv.Itoa(q.Length))
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

			ctx.Write(pq.Dequeue())
		case CmdPQLen:
			pq, found := PQueues[cmd.Key]
			if !found {
				break
			}

			ctx.Write(strconv.Itoa(pq.Length))
		}
	}
}

type Server struct {
	port   string
	ln     net.Listener
	quitCh chan struct{}

	// Auth
	user        string
	password    string
	requireAuth bool
}

func NewServer(port string) *Server {
	return &Server{
		port:   port,
		quitCh: make(chan struct{}),
	}
}

func (s *Server) Auth(user string, password string) {
	s.requireAuth = true
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
	for {
		line, err := ctx.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}

			fmt.Println("Error while reading data for connection", err)
			break
		}

		s.Execute(ctx, line)
		ctx.End()
	}

}

func main() {
	var (
		port        string
		portSr      string
		disableAuth bool
		user        string
		userSr      string
		password    string
		passwordSr  string
	)

	flag.StringVar(&port, "port", "", "Port to run server")
	flag.StringVar(&portSr, "p", "", "Shorthand for port")
	flag.BoolVar(&disableAuth, "noauth", false, "Disable authentication")
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
	if !disableAuth {
		server.Auth(user, password)
	}

	server.Start()
}
