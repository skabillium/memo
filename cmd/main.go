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
	"strings"
	"sync"
	"time"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

const DefaultUser = "memo"
const DefaultPassword = "password"
const CurrentRespVersion = "2"

const DefaultCleanupLimit = 20

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
		return s.db.CleanupExpired(cmd.Limit)
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
		return s.db.Del(cmd.Keys)
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
	case CmdSetAdd:
		added, err := s.db.SetAdd(cmd.Key, cmd.Values)
		if err != nil {
			return err
		}

		return added
	case CmdSetMembers:
		members, err := s.db.SetMembers(cmd.Key)
		if err != nil {
			return err
		}

		return members
	case CmdSetRem:
		removed, err := s.db.SetRemove(cmd.Key, cmd.Values)
		if err != nil {
			return err
		}
		return removed
	case CmdSetIsMember:
		ismember, err := s.db.SetIsMember(cmd.Key, cmd.Value)
		if err != nil {
			return err
		}
		return ismember
	case CmdSetCard:
		size, err := s.db.SetCard(cmd.Key)
		if err != nil {
			return err
		}
		return size
	case CmdSetInter:
		inter, err := s.db.SetInter(cmd.Keys[0], cmd.Keys[1])
		if err != nil {
			return err
		}
		return inter
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
	ln      net.Listener
	quitCh  chan struct{}
	options *ServerOptions
	db      *db.Database
	walch   chan string

	// Server info
	connMu sync.Mutex // Mutex to increment connections
	Info   ServerInfo
}

func NewServer(options *ServerOptions) *Server {
	return &Server{
		options: options,
		quitCh:  make(chan struct{}),
		db:      db.NewDatabase(),
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
	var ops int
	for {
		// WAL is serialized as bulk strings
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
	ln, err := net.Listen("tcp", "localhost:"+s.options.Port)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln
	go s.acceptLoop()

	if s.options.AutoCleanupEnabled {
		go s.runExpireJob()
		fmt.Println("Started auto cleanup job")
	}

	if s.options.WalEnabled {
		s.walch = make(chan string)
		fmt.Println("WAL enabled:", WalName)
		go writeToWAL(s.walch)
	}

	fmt.Println("Memo server started on port", s.options.Port)

	<-s.quitCh

	close(s.walch)

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

		if s.options.WalEnabled {
			s.walch <- exec
		}

		res := s.Execute(command)
		ctx.Write(res)
		ctx.End()
	}
}

func (s *Server) CanExecute(ctx *MemoContext, command *Command) error {
	if s.options.AuthEnabled && !ctx.hasAuth {
		if command.Kind != CmdHello && command.Kind != CmdAuth {
			return ErrNoAuth
		}

		if !(command.Auth.User == s.options.User && command.Auth.Password == s.options.Password) {
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

			if strings.Contains(s, " ") {
				s = "\"" + s + "\""
			}

			exec += s
		}
	default:
		return "", ErrUnsupportedType
	}

	return exec, nil
}

func (s *Server) runExpireJob() {
	// Remove expired keys every second
	ticker := time.NewTicker(s.options.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.db.CleanupExpired(s.options.CleanupLimit)
	}
}

type ServerOptions struct {
	Port               string
	AuthEnabled        bool
	WalEnabled         bool
	AutoCleanupEnabled bool
	CleanupLimit       int
	CleanupInterval    time.Duration
	User               string
	Password           string
}

func getServerOptions() *ServerOptions {
	var (
		port            string
		portSr          string
		disableAuth     bool
		enableWal       bool
		disableCleanup  bool
		cleanupLimit    int
		cleanupInterval int
		user            string
		userSr          string
		password        string
		passwordSr      string
	)

	flag.StringVar(&port, "port", "", "Port to run server")
	flag.StringVar(&portSr, "p", "", "Shorthand for port")
	flag.BoolVar(&disableAuth, "noauth", false, "Disable authentication")
	flag.BoolVar(&enableWal, "wal", false, "Enable write ahead log authentication")
	flag.BoolVar(&disableCleanup, "nocleanup", false, "Disable auto cleanup")
	flag.IntVar(&cleanupLimit, "cleanup-limit", 0, "Cleanup limit")
	flag.IntVar(&cleanupInterval, "cleanup-interval", 1, "Cleanup interval in seconds")
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

	if cleanupLimit == 0 {
		cleanupLimit = DefaultCleanupLimit
	}

	options := &ServerOptions{
		Port:               port,
		AuthEnabled:        !disableAuth,
		WalEnabled:         enableWal,
		AutoCleanupEnabled: !disableCleanup,
		CleanupLimit:       cleanupLimit,
		CleanupInterval:    time.Duration(cleanupInterval) * time.Second,
		User:               user,
		Password:           password,
	}

	return options
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

func main() {
	options := getServerOptions()
	server := NewServer(options)

	if options.WalEnabled {
		if FileExists(WalName) {
			ops, err := server.BuildDbFromWal()
			if err != nil {
				fmt.Println("Failed to initialize database from WAL")
				fmt.Println(err)
			} else {
				fmt.Printf("Initialized database from %d commands\n", ops)
			}
		}
	}

	log.Fatal(server.Start())
}
