package main

import (
	"context"
	"ieeeuottawa/vend-server/internal/router"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	r := router.NewRouter()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "db-redis:6379",
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	err := rdb.Set(ctx, "foo", "bar", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := rdb.Get(ctx, "foo").Result()
	if err != nil {
		panic(err)
	}

	log.Println("foo: ", val)

	log.Println("Server starting on port :3000")
	http.ListenAndServe(":3000", r)
}
