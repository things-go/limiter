package main

import (
	"context"
	"log"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func main() {
	var script = `
	local key = KEYS[1]

	local tb = {}
	local cnt = tonumber(redis.call("GET", key))
	if cnt == nil then 
		tb[1] = 0
		return tb
	end
	local ttl =  tonumber(redis.call("TTL", key))
	if ttl == nil then 
		tb[1] = 0
		return tb
	end
	tb[1] = 1
	tb[2] = cnt
	tb[3] = ttl
	return tb
	
	`
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatal(err)
	}
	defer mr.Close()
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	value, err := client.Eval(
		context.Background(),
		script,
		[]string{"abc"},
	).Result()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(value)
}
