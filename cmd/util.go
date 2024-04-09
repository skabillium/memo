package main

import (
	"errors"
	"flag"
	"os"
	"strings"
	"time"
)

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

// Read command line options
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

// Convert parsed request to a string
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

// Check if a given file path exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
