package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
var ErrUnsupportedType = errors.New("ERR unsupported type for request")

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

func (s *Server) Execute(cmd *Command) any {
	switch cmd.Kind {
	case CmdVersion:
		return resp.SimpleString(MemoVersion)
	case CmdPing:
		return resp.SimpleString("PONG")
	case CmdHello:
		if cmd.RespVersion != CurrentRespVersion {
			return ErrNoProto
		}
		return s.Info
	case CmdInfo:
		return "Memo server version " + MemoVersion
	case CmdKeys:
		keys := s.db.Keys(cmd.Pattern)
		return keys
	case CmdFlushAll:
		s.db.FlushAll()
		return resp.SimpleString("OK")
	case CmdDbSize:
		return s.db.Size()
	case CmdCleanup:
		deleted := s.db.CleanupExpired()
		return deleted
	case CmdExpire:
		ok := s.db.Expire(cmd.Key, cmd.ExpireIn)
		if !ok {
			return 0
		}
		return 1
	case CmdSet:
		s.db.Set(cmd.Key, cmd.Value, cmd.ExpireIn)
		return resp.SimpleString("OK")
	case CmdGet:
		value, found, err := s.db.Get(cmd.Key)
		if err != nil {
			return err
		}

		if !found {
			return nil
		}

		return value
	case CmdList:
		return errors.New("ERR unsupported command 'list'")
	case CmdDel:
		s.db.Del(cmd.Key)
		return resp.SimpleString("OK")
	case CmdQueueAdd:
		s.db.PQAdd(cmd.Key, cmd.Value, cmd.Priority)
		return 1
	case CmdQueuePop:
		value, found, err := s.db.PQPop(cmd.Key)
		if err != nil {
			return err
		}

		if !found {
			return nil
		}

		return value
	case CmdQueueLen:
		length, found, err := s.db.PQLen(cmd.Key)
		if err != nil {
			return err
		}

		if !found {
			return nil
		}

		return length
	case CmdLPush:
		s.db.LPush(cmd.Key, cmd.Values)
		return len(cmd.Values)
	case CmdRPush:
		s.db.RPush(cmd.Key, cmd.Values)
		return len(cmd.Values)
	case CmdLPop:
		value, found, err := s.db.LPop(cmd.Key)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		return value
	case CmdRPop:
		value, found, err := s.db.RPop(cmd.Key)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}
		return value
	case CmdLLen:
		length, err := s.db.LLen(cmd.Key)
		if err != nil {
			return err
		}
		return length
	}

	return nil
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

	db         *db.Database
	walEnabled bool
	wal        *WAL

	// Auth
	user        string
	password    string
	requireAuth bool
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

func (s *Server) EnableWal(wal *WAL) {
	s.walEnabled = true
	s.wal = wal
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

func (s *Server) BuildDbFromWal() (int, error) {
	file, err := os.Open(WalName)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	rd := bufio.NewReader(file)
	ops := 0
	for {
		line, err := resp.Read(rd)
		if err != nil {
			if err != io.EOF {
				return -1, err
			}
			break
		}

		exec, ok := line.(string)
		if !ok {
			s.db.FlushAll()
			return -1, errors.New("corrupted wal file, please manually verify that the contents are correct")
		}

		cmd, err := ParseCommand(exec)
		if err != nil {
			return -1, err
		}

		s.Execute(cmd)
		ops++
	}

	return ops, nil
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", "localhost:"+s.port)
	if err != nil {
		return err
	}
	defer ln.Close()

	fmt.Println("Memo server started on port", s.port)
	if s.walEnabled {
		fmt.Println("WAL enabled:", WalName)
	}

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
			continue
		}

		command, err := ParseCommand(exec)
		if err != nil {
			ctx.EndWith(err)
			continue
		}

		if err = s.CanExecute(ctx, command); err != nil {
			ctx.EndWith(err)
			continue
		}

		if s.walEnabled {
			s.wal.Write(exec)
		}

		res := s.Execute(command)
		ctx.Write(res)
		ctx.End()
	}
}

func (s *Server) CanExecute(ctx *MemoContext, command *Command) error {
	if s.requireAuth && !ctx.hasAuth {
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
				return "", ErrUnsupportedType
			}
			if i != 0 {
				exec += " "
			}
			exec += s
		}
	default:
		return "", ErrUnsupportedType
	}

	return exec, nil
}

type ServerOptions struct {
	Port       string
	EnableAuth bool
	EnableWal  bool
	User       string
	Password   string
}

func getServerOptions() *ServerOptions {
	var (
		port        string
		portSr      string
		disableAuth bool
		enableWal   bool
		user        string
		userSr      string
		password    string
		passwordSr  string
	)

	flag.StringVar(&port, "port", "", "Port to run server")
	flag.StringVar(&portSr, "p", "", "Shorthand for port")
	flag.BoolVar(&disableAuth, "noauth", false, "Disable authentication")
	flag.BoolVar(&enableWal, "wal", false, "Enable write ahead log authentication")
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
		EnableWal:  enableWal,
		User:       user,
		Password:   password,
	}

	return options
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func main() {
	options := getServerOptions()
	server := NewServer(options.Port)
	if options.EnableAuth {
		server.Auth(options.User, options.Password)
	}

	if options.EnableWal {
		if FileExists(WalName) {
			ops, err := server.BuildDbFromWal()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Failed to initialize database from WAL")
			} else {
				fmt.Printf("Initialized database from %d commands\n", ops)
			}

		}

		wal, err := NewWal()
		if err != nil {
			log.Fatal("Could not initialize wal")
		} else {
			defer wal.Close()
			server.EnableWal(wal)
		}

	}

	log.Fatal(server.Start())
}
