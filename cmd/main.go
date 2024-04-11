package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"skabillium/memo/cmd/db"
	"skabillium/memo/cmd/resp"
	"sync"
	"time"
)

const MemoVersion = "0.0.1"
const DefaultPort = "5678"

const DefaultUser = "memo"
const DefaultPassword = "password"
const CurrentRespVersion = "2"

// Number of expired keys to remove every cleanup cycle
// see: runExpireJob() and db.CleanupExpired()
const DefaultCleanupLimit = 20

var ErrNoAuth = errors.New("NOAUTH Authentication required")
var ErrWrongPass = errors.New("WRONGPASS invalid username-password pair or user is disabled")
var ErrNoProto = errors.New("NOPROTO unsupported protocol version")
var ErrUnsupportedType = errors.New("ERR unsupported type for request")

// Helper struct for holding any connection specific information and communicating with
// clients
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

func (c *MemoContext) EndWith(message any) {
	c.Write(message)
	c.End()
}

func (c *MemoContext) End() {
	c.rw.Flush()
}

// Run a given command against the database
func (s *Server) Execute(cmd *Command) any {
	s.dbmu.Lock()
	defer s.dbmu.Unlock()

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
		s.db.PQAdd(cmd.Key, cmd.Values, cmd.Priority)
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
	dbmu    sync.Mutex // Mutex to synchronize db acces from different connections
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

// Increment server connections count
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

// If the Write Ahead Log is enabled, this function reads it and rebuilds the database from the
// logged commands. It is meant to run only once, before the server is actually started.
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

// Start server with the options provided by the cli
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

// Accept new connections
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

// This executes when a new connection is created, it runs in a separate goroutine for
// every connection and has it's own separate MemoContext.
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

		if command.Kind == CmdQuit {
			break
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

// If authentication is enabled, this function checks if the command provided is the "hello"
// or "auth" commands (the only commands that can perform authentication) and if the credentials
// provided are correct.
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

// Job to delete expired keys, by default it runs every second but can be customized
// with the "cleanup-interval" cli option (in seconds)
func (s *Server) runExpireJob() {
	// Remove expired keys every second
	ticker := time.NewTicker(s.options.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		s.dbmu.Lock()
		s.db.CleanupExpired(s.options.CleanupLimit)

		s.dbmu.Unlock()
	}
}

// The actual main function of the server, it reads and options provided
// it instantiates the server and initialized the database
func runServer() error {
	flag.Usage = func() {
		fmt.Println("Memo server version", MemoVersion)
		fmt.Println("Available options:")
		flag.PrintDefaults()
	}

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

	return server.Start()
}

// Main is only responsible for running the runServer function and exiting with an error
// if it occurs
func main() {
	err := runServer()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
