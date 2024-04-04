package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"skabillium/memo/cmd/resp"
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
	payload, err := resp.Serialize(message)
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

func (c *MemoContext) Simple(message string) {
	c.rw.WriteString(resp.SerializeSimpleStr(message))
}

func (c *MemoContext) Ok() {
	c.Simple("OK")
}

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

		hello := commands[0]
		if hello.Kind != CmdHello {
			ctx.Error(errors.New("authentication required"))
			return
		}

		if hello.RespVersion != CurrentRespVersion {
			ctx.Error(errors.New("NOPROTO unsupported protocol version"))
			return
		}

		if !(hello.Auth.User == s.user && hello.Auth.Password == s.password) {
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
			ctx.Simple("PONG")
		case CmdHello:
			ctx.Write(s.Info)
		case CmdInfo:
			ctx.Write("Memo server version " + MemoVersion)
		case CmdKeys:
			keys := s.db.Keys()
			ctx.Write(keys)
		case CmdFlushAll:
			s.db.FlushAll()
			ctx.Ok()
		case CmdSet:
			s.db.Set(cmd.Key, cmd.Value)
			ctx.Ok()
		case CmdGet:
			store, found := s.db.Get(cmd.Key)
			if !found {
				ctx.Write(nil)
				break
			}

			ctx.Write(store.Value)
		case CmdList:
			ctx.Error(errors.New("ERR unsupported command 'list'"))
		case CmdDel:
			s.db.Del(cmd.Key)
			ctx.Ok()
		case CmdQueueAdd:
			s.db.Qadd(cmd.Key, cmd.Value)
			ctx.Ok()
		case CmdQueuePop:
			value, found, err := s.db.QPop(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}

			if !found {
				ctx.Write(nil)
				break
			}

			ctx.Write(value)
		case CmdQueueLen:
			length, found, err := s.db.Qlen(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}

			if !found {
				ctx.Write(nil)
			}

			ctx.Write(length)
		case CmdPQAdd:
			s.db.PQAdd(cmd.Key, cmd.Value, cmd.Priority)
			ctx.Ok()
		case CmdPQPop:
			value, found, err := s.db.PQPop(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}

			if !found {
				ctx.Write(nil)
				break
			}

			ctx.Write(value)
		case CmdPQLen:
			length, found, err := s.db.PQLen(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}

			if !found {
				ctx.Write(nil)
				break
			}

			ctx.Write(length)
		case CmdLPush:
			s.db.LPush(cmd.Key, cmd.Value)
			ctx.Write(1)
		case CmdRPush:
			s.db.RPush(cmd.Key, cmd.Value)
			ctx.Write(1)
		case CmdLPop:
			value, found, err := s.db.LPop(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}
			if !found {
				ctx.Write(nil)
				break
			}
			ctx.Write(value)
		case CmdRPop:
			value, found, err := s.db.RPop(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}
			if !found {
				ctx.Write(nil)
				break
			}
			ctx.Write(value)
		case CmdLLen:
			length, err := s.db.LLen(cmd.Key)
			if err != nil {
				ctx.Error(err)
				break
			}
			ctx.Write(length)
		}
	}
}

type ServerInfo struct {
	Server  string
	Version string
	Proto   int
	Mode    string
	Modules []string
}

type Server struct {
	port   string
	ln     net.Listener
	quitCh chan struct{}
	// Auth
	user        string
	password    string
	requireAuth bool
	// Db
	db *Database
	// Server info
	Info ServerInfo
}

func NewServer(port string) *Server {
	return &Server{
		port:   port,
		quitCh: make(chan struct{}),
		db:     NewDatabase(),
		Info: ServerInfo{
			Server:  "memo",
			Version: MemoVersion,
			Proto:   2,
			Mode:    "standalone",
			Modules: []string{},
		},
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

	ctx := NewMemoContext(conn)
	for {
		req, err := resp.Read(ctx.rw.Reader)
		if err != nil {
			if err != io.EOF {
				fmt.Print(err)
			}
			break
		}

		var exec string
		switch req := req.(type) {
		case string:
			exec = req
		case []any:
			for i, v := range req {
				s, ok := v.(string)
				if !ok {
					fmt.Println("Could not cast", v, "to string")
					break
				}
				if i != 0 {
					exec += " "
				}
				exec += s
			}
		default:
			ctx.Error(errors.New("unsupported type for request"))
		}

		s.Execute(ctx, exec)
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
