package main

import (
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestKV(t *testing.T) {
	memo := GetClient()

	memo.Set(ctx, "name", "bill", 0).Err()
	memo.Set(ctx, "message", "hello there!", 1*time.Second).Err()

	name, err := memo.Get(ctx, "name").Result()
	if name != "bill" || err != nil {
		t.Error("Expected Get('name') to return 'bill'")
	}

	message, err := memo.Get(ctx, "message").Result()
	if message != "hello there!" || err != nil {
		t.Error("Expected Get('message') to return 'hello there!'")
	}

	time.Sleep(1 * time.Second)

	err = memo.Get(ctx, "message").Err()
	if err != redis.Nil {
		t.Error("Expected message to be expired")
	}

	memo.FlushAll(ctx) // Cleanup
}
