package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"skabillium/memo/cmd/db"
	"skabillium/memo/cmd/resp"
	"sync"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

const DefaultUser = "memo"
const DefaultPassword = "password"
const CurrentRespVersion = "2"

var ErrNoAuth = errors.New("NOAUTH Authentication required")
var ErrWrongPass = errors.New("WRONGPASS invalid username-password pair or user is disabled")
var ErrNoProto = errors.New("NOPROTO unsupported protocol version")

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

func (s *Server) Execute(ctx *MemoContext, cmd *Command) {
	switch cmd.Kind {
	case CmdVersion:
		ctx.Write(MemoVersion)
	case CmdPing:
		ctx.Simple("PONG")
	case CmdHello:
		if cmd.RespVersion != CurrentRespVersion {
			ctx.Error(ErrNoProto)
			break
		}
		ctx.Write(s.Info)
	case CmdInfo:
		ctx.Write("Memo server version " + MemoVersion)
	case CmdKeys:
		keys := s.db.Keys()
		ctx.Write(keys)
	case CmdFlushAll:
		s.db.FlushAll()
		ctx.Ok()
	case CmdCleanup:
		deleted := s.db.CleanupExpired()
		ctx.Write(deleted)
	case CmdExpire:
		ok := s.db.Expire(cmd.Key, cmd.ExpireIn)
		if !ok {
			ctx.Write(0)
			break
		}
		ctx.Write(1)
	case CmdSet:
		s.db.Set(cmd.Key, cmd.Value, cmd.ExpireIn)
		ctx.Ok()
	case CmdGet:
		value, found, err := s.db.Get(cmd.Key)
		if err != nil {
			ctx.Error(err)
			break
		}

		if !found {
			ctx.Write(nil)
			break
		}

		ctx.Write(value)
	case CmdList:
		ctx.Error(errors.New("ERR unsupported command 'list'"))
	case CmdDel:
		s.db.Del(cmd.Key)
		ctx.Ok()
	case CmdQueueAdd:
		s.db.PQAdd(cmd.Key, cmd.Value, cmd.Priority)
		ctx.Write(1)
	case CmdQueuePop:
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
	case CmdQueueLen:
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
		s.db.LPush(cmd.Key, cmd.Values)
		ctx.Write(len(cmd.Values))
	case CmdRPush:
		s.db.RPush(cmd.Key, cmd.Values)
		ctx.Write(len(cmd.Values))
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

type ServerInfo struct {
	Server      string
	Version     string
	Proto       int
	Mode        string
	Modules     []string
	Connections int
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
	db *db.Database
	// Server info
	connMu sync.Mutex // Mutex to increment connections
	Info   ServerInfo
}

func NewServer(port string) *Server {
	return &Server{
		port:   port,
		quitCh: make(chan struct{}),
		db:     db.NewDatabase(),
		Info: ServerInfo{
			Server:      "memo",
			Version:     MemoVersion,
			Proto:       2,
			Mode:        "standalone",
			Modules:     []string{},
			Connections: 0,
		},
	}
}

func (s *Server) Auth(user string, password string) {
	s.requireAuth = true
	s.user = user
	s.password = password
}

func (s *Server) newConn() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.Info.Connections++
}

func (s *Server) closeConn(conn net.Conn) {
	conn.Close()

	s.connMu.Lock()
	defer s.connMu.Unlock()
	s.Info.Connections--
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

		s.newConn()
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.closeConn(conn)

	ctx := NewMemoContext(conn)
	for {
		req, err := resp.Read(ctx.rw.Reader)
		if err != nil {
			if err != io.EOF {
				fmt.Print(err)
			}
			break
		}

		exec, err := StringifyRequest(req)
		if err != nil {
			ctx.EndWith(err)
			break
		}

		command, err := ParseCommand(exec)
		if err != nil {
			ctx.EndWith(err)
			break
		}

		// TODO: Maybe use pointers to arrays instead of copying arrays left and right
		if err = s.CanExecute(ctx, command); err != nil {
			ctx.EndWith(err)
			break
		}

		s.Execute(ctx, command)
		ctx.End()
	}
}

func (s *Server) CanExecute(ctx *MemoContext, command *Command) error {
	if !ctx.hasAuth && s.requireAuth {
		if command.Kind != CmdHello && command.Kind != CmdAuth {
			return ErrNoAuth
		}

		if !(command.Auth.User == s.user && command.Auth.Password == s.password) {
			return ErrWrongPass
		}

		ctx.Authenticate()
	}

	return nil
}

func StringifyRequest(req any) (string, error) {
	var exec string
	switch req := req.(type) {
	case string:
		exec = req
	case []any:
		for i, v := range req {
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("could not cast '%v' to string", v)
			}
			if i != 0 {
				exec += " "
			}
			exec += s
		}
	default:
		return "", errors.New("unsupported type for request")
	}

	return exec, nil
}

type ServerOptions struct {
	Port       string
	EnableAuth bool
	User       string
	Password   string
}

func getServerOptions() *ServerOptions {
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

	options := &ServerOptions{
		Port:       port,
		EnableAuth: !disableAuth,
		User:       user,
		Password:   password,
	}

	return options
}

func main() {
	options := getServerOptions()
	server := NewServer(options.Port)
	if options.EnableAuth {
		server.Auth(options.User, options.Password)
	}

	server.Start()
}
