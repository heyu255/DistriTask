package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/heyu255/distritask/internal/queue"
	"github.com/heyu255/distritask/internal/task"
	"github.com/redis/go-redis/v9"
)

func main() {
	// 1. Setup Redis Connection
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Test Redis connection on startup
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Fatal: Could not connect to Redis: %v", err)
	}

	// 2. Initialize the Queue
	taskQueue := queue.NewRedisQueue(rdb, "task_stream")

	// 3. Create a new ServeMux to handle routes
	mux := http.NewServeMux()

	// 4. Define the /submit handler
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Create the task object
		t := &task.Task{
			ID:        uuid.New().String(),
			Name:      "ExampleTask",
			Status:    task.Pending,
			CreatedAt: time.Now(),
		}

		// Push to Redis
		err := taskQueue.Enqueue(context.Background(), t)
		if err != nil {
			log.Printf("[MANAGER] Enqueue error: %v", err)
			http.Error(w, "Internal error", 500)
			return
		}

		// Broadcast initial "pending" status so frontend shows it immediately
		broadcastStatus(context.Background(), rdb, t.ID, "pending", "manager", "Task accepted and waiting in queue")

		// Log success to terminal
		log.Printf("[MANAGER] Successfully enqueued task: %s", t.ID)

		// Respond to Frontend
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": "%s", "message": "Task enqueued successfully!"}`, t.ID)
	})

	// 5. Wrap the mux with the CORS middleware and start the server
	log.Println("Manager starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", enableCORS(mux)))
}

// enableCORS middleware handles the pre-flight OPTIONS request and sets headers
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your Next.js frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
