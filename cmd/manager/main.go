package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/heyu255/distritask/internal/queue"
	"github.com/heyu255/distritask/internal/task"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Setup Redis Connection
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

	// Test Redis connection on startup
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Fatal: Could not connect to Redis: %v", err)
	}

	// 2. Initialize the Queue
	taskQueue := queue.NewRedisQueue(rdb, "task_stream")

	// 3. Create a new ServeMux to handle routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	// 4. Define the /submit handler with panic recovery
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Recover from any panics
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[MANAGER] Panic recovered: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		log.Printf("[MANAGER] Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		
		// Only allow POST method
		if r.Method != "POST" {
			log.Printf("[MANAGER] Method not allowed: %s (expected POST)", r.Method)
			http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
			return
		}
		
		// Create the task object
		t := &task.Task{
			ID:        uuid.New().String(),
			Name:      "ExampleTask",
			Status:    task.Pending,
			CreatedAt: time.Now(),
		}

		log.Printf("[MANAGER] Created task: %s", t.ID)

		// Push to Redis
		err := taskQueue.Enqueue(context.Background(), t)
		if err != nil {
			log.Printf("[MANAGER] Enqueue error: %v", err)
			http.Error(w, "Internal error", 500)
			return
		}

		log.Printf("[MANAGER] Task enqueued successfully: %s", t.ID)

		// Broadcast initial "pending" status so frontend shows it immediately
		broadcastStatus(context.Background(), rdb, t.ID, "pending", "manager", "Task accepted and waiting in queue")

		// Log success to terminal
		log.Printf("[MANAGER] Successfully enqueued task: %s", t.ID)

		// Respond to Frontend
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": "%s", "message": "Task enqueued successfully!"}`, t.ID)
		log.Printf("[MANAGER] Response sent for task: %s", t.ID)
	})

	// 5. Wrap the mux with the CORS middleware and start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}
	log.Printf("Manager starting on :%s...", port)
	log.Fatal(http.ListenAndServe(":"+port, enableCORS(mux)))
}

// enableCORS middleware handles the pre-flight OPTIONS request and sets headers
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your Next.js frontend origin (use env var in production)
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000" // fallback for local dev
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle browser pre-flight request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// broadcastStatus broadcasts task status updates to Redis Pub/Sub
func broadcastStatus(ctx context.Context, rdb *redis.Client, taskID, status, worker, message string) {
	update := fmt.Sprintf(`{"id": "%s", "status": "%s", "time": "%s", "worker": "%s", "message": "%s"}`,
		taskID, status, time.Now().Format("15:04:05"), worker, message)

	rdb.Publish(ctx, "task_updates", update)
}
