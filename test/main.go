package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func GetClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:5678",
	})
}

func main() {
	memo := GetClient()

	memo.Set(ctx, "name", "bill", 1*time.Second)

	time.Sleep(2 * time.Second)

	err := memo.Get(ctx, "name").Err()
	if err != nil {
		fmt.Println(err == redis.Nil)
	}
}
