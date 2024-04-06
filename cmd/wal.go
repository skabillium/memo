package main

import (
	"fmt"
	"os"
	"skabillium/memo/cmd/resp"
)

const WalName = "wal.log"

func writeToWAL(walch <-chan string) {
	file, err := os.OpenFile(WalName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	for line := range walch {
		_, err := file.WriteString(resp.SerializeStr(line))
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
