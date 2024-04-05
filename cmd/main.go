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

const WalName = "wal.log"

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

func (s *Server) Execute(cmd *Command) (any, bool) {
	switch cmd.Kind {
	case CmdVersion:
		return MemoVersion, true
	case CmdPing:
		return "PONG", true
	case CmdHello:
		if cmd.RespVersion != CurrentRespVersion {
			return ErrNoProto, false
		}
		return s.Info, false
	case CmdInfo:
		return "Memo server version " + MemoVersion, false
	case CmdKeys:
		keys := s.db.Keys()
		return keys, false
	case CmdFlushAll:
		s.db.FlushAll()
		return "OK", true
	case CmdCleanup:
		deleted := s.db.CleanupExpired()
		return deleted, false
	case CmdExpire:
		ok := s.db.Expire(cmd.Key, cmd.ExpireIn)
		if !ok {
			return 0, false
		}
		return 1, false
	case CmdSet:
		s.db.Set(cmd.Key, cmd.Value, cmd.ExpireIn)
		return "OK", true
	case CmdGet:
		value, found, err := s.db.Get(cmd.Key)
		if err != nil {
			return err, false
		}

		if !found {
			return nil, false
		}

		return value, false
	case CmdList:
		return errors.New("ERR unsupported command 'list'"), false
	case CmdDel:
		s.db.Del(cmd.Key)
		return "OK", true
	case CmdQueueAdd:
		s.db.PQAdd(cmd.Key, cmd.Value, cmd.Priority)
		return 1, false
	case CmdQueuePop:
		value, found, err := s.db.PQPop(cmd.Key)
		if err != nil {
			return err, false
		}

		if !found {
			return nil, false
		}

		return value, false
	case CmdQueueLen:
		length, found, err := s.db.PQLen(cmd.Key)
		if err != nil {
			return err, false
		}

		if !found {
			return nil, false
		}

		return length, false
	case CmdLPush:
		s.db.LPush(cmd.Key, cmd.Values)
		return len(cmd.Values), false
	case CmdRPush:
		s.db.RPush(cmd.Key, cmd.Values)
		return len(cmd.Values), false
	case CmdLPop:
		value, found, err := s.db.LPop(cmd.Key)
		if err != nil {
			return err, false
		}
		if !found {
			return nil, false
		}
		return value, false
	case CmdRPop:
		value, found, err := s.db.RPop(cmd.Key)
		if err != nil {
			return err, false
		}
		if !found {
			return nil, false
		}
		return value, false
	case CmdLLen:
		length, err := s.db.LLen(cmd.Key)
		if err != nil {
			return err, false
		}
		return length, false
	}

	return nil, false
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

	scanner := bufio.NewScanner(file)
	ops := 0
	for scanner.Scan() {
		line := scanner.Text()
		command, err := ParseCommand(line)
		if err != nil {
			return -1, fmt.Errorf("Error while executing command: '%s'", line)
		}

		s.Execute(command)
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
			s.wal.Writeln(exec)
		}

		res, simple := s.Execute(command)
		if simple {
			ctx.Simple(res.(string))
		} else {
			ctx.Write(res)
		}
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
				log.Fatal(err)
			}

			fmt.Printf("Built database from %d commands\n", ops)
		}

		wal, err := NewWal()
		if err != nil {
			// TODO: Handle this differently
			log.Fatal("Could not initialize wal")
		}
		defer wal.Close()
		server.EnableWal(wal)
	}

	server.Start()
}
