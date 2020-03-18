package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	// "github.com/bsm/redislock"
	"github.com/go-redis/redis/v7"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from SSP")
}

func fooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func newRedisClient() {
	client := redis.NewClient(&redis.Options{
		Addr:     "redisssp:6379",
		Password: "",
		DB:       0,
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}

func main() {
	newRedisClient()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/foo", fooHandler)

	s := &http.Server{
		Addr:           ":6060",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Fatal(s.ListenAndServe())
}
