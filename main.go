package main

import (
	"log"
	"os"

	"github.com/garyburd/redigo/redis"
)

func main() {
	var (
		redisURL = mustGetenv("REDISCLOUD_URL")
		//accessToken = mustGetenv("ACCESS_TOKEN")
	)

	conn, err := redis.Dial("tcp", redisURL)
	if err != nil {
		log.Println("Error:", err)
	}
	defer conn.Close()
}

func mustGetenv(key string) (value string) {
	v := os.Getenv(key)
	if v == "" {
		panic(key + " is not set")
	}
	return v
}
