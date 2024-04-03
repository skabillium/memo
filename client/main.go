package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:5678")
	if err != nil {
		log.Fatal(err)
	}

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	// resp := "*1\r\n$7\r\nhello 3\r\n"

	set := "set name bill\n"
	// get := "get name\n"

	rw.WriteString(set)
	rw.Flush()

	line, err := rw.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(line)
}
