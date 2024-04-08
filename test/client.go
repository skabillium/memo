package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func GetClient() *redis.Client {
	memo := redis.NewClient(&redis.Options{
		Addr: "localhost:5678",
	})

	err := memo.Ping(ctx).Err()
	if err != nil {
		fmt.Println("Could not connect to Memo server, make sure it is running")
		fmt.Println(err)
		os.Exit(1)
	}

	return memo
}
