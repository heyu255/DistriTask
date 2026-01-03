package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	redisURL := os.Getenv("REDIS_URL")
	// Remove quotes if present (Railway sometimes adds them)
	redisURL = strings.Trim(redisURL, `"'`)
	var rdb *redis.Client
	if redisURL == "" {
		// Fallback for local dev
		rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	} else {
		// Parse full Redis URL (handles rediss://, redis://, etc.)
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("Failed to parse REDIS_URL: %v", err)
		}
		rdb = redis.NewClient(opt)
	}
	ctx := context.Background()

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}
		defer conn.Close()

		// Subscribe to Redis Pub/Sub
		pubsub := rdb.Subscribe(ctx, "task_updates")
		defer pubsub.Close()
		ch := pubsub.Channel()

		for msg := range ch {
			// Push every Redis update straight to the browser
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // fallback for local dev
	}
	log.Printf("Dashboard WebSocket Server on :%s...", port)
	http.ListenAndServe(":"+port, nil)
}
