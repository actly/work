package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/liu1700/work/webui"
)

var (
	redisHost      = flag.String("host", "127.0.0.1", "redis host")
	redisPort      = flag.String("port", "6379", "redis port")
	redisDatabase  = flag.String("database", "0", "redis database")
	redisNamespace = flag.String("namespace", "work", "redis namespace")
	webHostPort    = flag.String("listen", ":5040", "hostport to listen for HTTP JSON API")
)

func main() {
	flag.Parse()

	fmt.Println("Starting workwebui:")
	fmt.Println("redis host = ", *redisHost)
	fmt.Println("redis port = ", *redisPort)
	fmt.Println("database = ", *redisDatabase)
	fmt.Println("namespace = ", *redisNamespace)
	fmt.Println("listen = ", *webHostPort)

	database, err := strconv.Atoi(*redisDatabase)
	if err != nil {
		fmt.Printf("Error: %v is not a valid database value", *redisDatabase)
		return
	}

	pool := newPool(*redisHost, *redisPort, database)

	server := webui.NewServer(*redisNamespace, pool, *webHostPort)
	server.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)

	<-c

	fmt.Println("Stopping server...")
	server.Stop()

	fmt.Println("\nQuitting...")
}

func newPool(addr string, port string, database int) *redis.Pool {
	return &redis.Pool{
		MaxActive:   3,
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%s", addr, port), redis.DialDatabase(database))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
