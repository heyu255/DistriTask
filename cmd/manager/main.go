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
	// Force immediate log output (flush stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	
	// Immediate test to verify binary is running
	fmt.Fprintf(os.Stderr, "[MANAGER] Binary started successfully\n")
	fmt.Fprintf(os.Stdout, "[MANAGER] Binary started successfully\n")
	
	log.Printf("[MANAGER] ========================================")
	log.Printf("[MANAGER] Starting Manager service...")
	log.Printf("[MANAGER] ========================================")
	
	// Log environment info
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[MANAGER] PORT environment variable: %s", port)
	log.Printf("[MANAGER] ALLOWED_ORIGIN: %s", os.Getenv("ALLOWED_ORIGIN"))
	
	// 1. Setup Redis Connection
	redisURL := os.Getenv("REDIS_URL")
	// Remove quotes if present (Railway sometimes adds them)
	redisURL = strings.Trim(redisURL, `"'`)
	log.Printf("[MANAGER] Redis URL configured: %s", maskRedisURL(redisURL))
	
	var rdb *redis.Client
	if redisURL == "" {
		log.Printf("[MANAGER] WARNING: No REDIS_URL found, using localhost fallback")
		// Fallback for local dev
		rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	} else {
		// Parse full Redis URL (handles rediss://, redis://, etc.)
		log.Printf("[MANAGER] Parsing Redis URL...")
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Fatalf("[MANAGER] FATAL: Failed to parse REDIS_URL: %v", err)
		}
		rdb = redis.NewClient(opt)
		log.Printf("[MANAGER] Redis client created successfully")
	}

	// Test Redis connection on startup
	log.Printf("[MANAGER] Testing Redis connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("[MANAGER] FATAL: Could not connect to Redis: %v", err)
	}
	log.Printf("[MANAGER] ✓ Redis connection successful")

	// 2. Initialize the Queue
	taskQueue := queue.NewRedisQueue(rdb, "task_stream")

	// 3. Create a new ServeMux to handle routes
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Manager service is running")
	})

	// 4. Define the /submit handler with panic recovery
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// Recover from any panics to prevent service crash
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
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		log.Printf("[MANAGER] Task enqueued to Redis: %s", t.ID)

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
	log.Printf("[MANAGER] ========================================")
	log.Printf("[MANAGER] Starting HTTP server on port :%s", port)
	log.Printf("[MANAGER] Health check: http://0.0.0.0:%s/", port)
	log.Printf("[MANAGER] Submit endpoint: http://0.0.0.0:%s/submit", port)
	log.Printf("[MANAGER] ========================================")
	
	// Use 0.0.0.0 to bind to all interfaces (required for Railway)
	addr := "0.0.0.0:" + port
	log.Printf("[MANAGER] Binding to: %s", addr)
	
	if err := http.ListenAndServe(addr, enableCORS(mux)); err != nil {
		log.Fatalf("[MANAGER] FATAL: HTTP server failed: %v", err)
	}
}

// enableCORS middleware handles the pre-flight OPTIONS request and sets headers
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow your Next.js frontend origin (use env var in production)
		allowedOrigin := os.Getenv("ALLOWED_ORIGIN")
		if allowedOrigin == "" {
			allowedOrigin = "http://localhost:3000" // fallback for local dev
		}
		
		log.Printf("[MANAGER] CORS: Allowing origin: %s, Request from: %s", allowedOrigin, r.Header.Get("Origin"))
		
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle browser pre-flight request
		if r.Method == "OPTIONS" {
			log.Printf("[MANAGER] CORS: Handling OPTIONS preflight request")
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

// maskRedisURL masks sensitive parts of Redis URL for logging
func maskRedisURL(url string) string {
	if url == "" {
		return "(empty)"
	}
	if len(url) > 20 {
		return url[:10] + "..." + url[len(url)-10:]
	}
	return url
}
