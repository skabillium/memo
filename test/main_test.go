package main

import (
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestServerCommands(t *testing.T) {
	memo := GetClient()

	info, err := memo.Info(ctx).Result()
	if err != nil || info != "Memo server version 0.0.1" {
		t.Error("Expected other response for Info()", info)
	}

	err = memo.Ping(ctx).Err()
	if err != nil {
		t.Error("Unexpected error with Ping()", err)
	}

	res, err := memo.Do(ctx, "version").Result()
	if err != nil {
		t.Error("Unexpected error", err)
	}
	version, ok := res.(string)
	if !ok {
		t.Error("Expected result of version to be string")
	}
	if version != "0.0.1" {
		t.Error("Expected version to be 0.0.1, got", version)
	}
}

func TestKV(t *testing.T) {
	memo := GetClient()
	defer memo.FlushAll(ctx)

	memo.Set(ctx, "name", "bill", 0).Err()
	memo.Set(ctx, "message", "hello there!", 1*time.Second).Err()
	memo.Set(ctx, "to_expire", "1", 0).Err()

	name, err := memo.Get(ctx, "name").Result()
	if name != "bill" || err != nil {
		t.Error("Expected Get('name') to return 'bill'")
	}

	err = memo.Del(ctx, "name").Err()
	if err != nil {
		t.Error("Error while deleting key 'name'", err)
	}

	err = memo.Get(ctx, "name").Err()
	if err != redis.Nil {
		t.Error("Expected Get('name') to return nil")
	}

	message, err := memo.Get(ctx, "message").Result()
	if message != "hello there!" || err != nil {
		t.Error("Expected Get('message') to return 'hello there!'")
	}

	err = memo.Expire(ctx, "to_expire", 1*time.Second).Err()
	if err != nil {
		t.Error("Unexpected error with Expire()", err)
	}

	time.Sleep(2 * time.Second)

	err = memo.Get(ctx, "message").Err()
	if err != redis.Nil {
		t.Error("Expected message to be expired")
	}
	err = memo.Get(ctx, "to_expire").Err()
	if err != redis.Nil {
		t.Error("Expected to_expire to be expired")
	}

}

func TestList(t *testing.T) {
	memo := GetClient()
	defer memo.FlushAll(ctx)

	list := "numbers"

	memo.RPush(ctx, list, "two", "three", "four")
	memo.LPush(ctx, list, "one")

	one, err := memo.LPop(ctx, list).Result()
	if err != nil {
		t.Error(err)
	}
	if one != "one" {
		t.Error("Expected one but received", one)
	}

	four, err := memo.RPop(ctx, list).Result()
	if err != nil {
		t.Error(err)
	}

	if four != "four" {
		t.Error("Expected four but received", four)
	}

	length, err := memo.LLen(ctx, list).Result()
	if err != nil {
		t.Error(err)
	}

	if length != 2 {
		t.Error("Expected length to be 2")
	}
}

func TestQueue(t *testing.T) {
	memo := GetClient()
	defer memo.FlushAll(ctx)

	queue := "names"
	memo.Do(ctx, "qadd", queue, "bill")
	memo.Do(ctx, "qadd", queue, "susan")
	memo.Do(ctx, "qadd", queue, "john", "2")
	memo.Do(ctx, "qadd", queue, "james")

	bill, err := memo.Do(ctx, "qpop", queue).Result()
	if err != nil {
		t.Error(err)
	}
	if bill != "bill" {
		t.Error("Expected item to be 'bill', got", bill)
	}

	susan, err := memo.Do(ctx, "qpop", queue).Result()
	if err != nil {
		t.Error(err)
	}
	if susan != "susan" {
		t.Error("Expected item to be 'susan', got", susan)
	}

	james, err := memo.Do(ctx, "qpop", queue).Result()
	if err != nil {
		t.Error(err)
	}
	if james != "james" {
		t.Error("Expected item to be 'james', got", james)
	}

	length, err := memo.Do(ctx, "qlen", queue).Result()
	if err != nil {
		t.Error(err)
	}
	if length.(int64) != 1 {
		t.Error("Expected length to be 1, got", length)
	}

	john, err := memo.Do(ctx, "qpop", queue).Result()
	if err != nil {
		t.Error(err)
	}
	if john != "john" {
		t.Error("Expected item to be 'john', got", john)
	}
}

func TestSet(t *testing.T) {
	memo := GetClient()
	defer memo.FlushAll(ctx)

	heroes := "heroes"
	marvel := "marvel"

	memo.SAdd(ctx, heroes, "superman", "batman", "wonder woman", "spiderman", "iron man")
	memo.SAdd(ctx, marvel, "spiderman", "iron man", "nick fury")

	heroesSize, err := memo.SCard(ctx, heroes).Result()
	if err != nil {
		t.Error(err)
	}

	if heroesSize != 5 {
		t.Error("Expected heroes cardinality to be 5, got", heroesSize)
	}

	ismem, err := memo.SIsMember(ctx, heroes, "batman").Result()
	if err != nil {
		t.Error(err)
	}

	if !ismem {
		t.Error("Expected 'batman' to be member of heroes")
	}

	ismem, err = memo.SIsMember(ctx, heroes, "nick fury").Result()
	if err != nil {
		t.Error(err)
	}

	if ismem {
		t.Error("Expected 'nick fury' to not be member of heroes")
	}

	inter, err := memo.SInter(ctx, heroes, marvel).Result()
	if err != nil {
		t.Error(err)
	}

	marvelHeroes := []string{"spiderman", "iron man"}
	if !reflect.DeepEqual(inter, marvelHeroes) {
		t.Error("Expected inter to be", marvelHeroes, "got", inter)
	}

	memo.SRem(ctx, heroes, "spiderman")

	heroesSize, err = memo.SCard(ctx, heroes).Result()
	if err != nil {
		t.Error(err)
	}

	if heroesSize != 4 {
		t.Error("Expected heroes cardinality to be 4, got", heroesSize)
	}
}
