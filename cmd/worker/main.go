package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/heyu255/distritask/internal/task"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	stream := "task_stream"
	group := "worker_group"

	// Ensure group exists (ignore error if group already exists)
	rdb.XGroupCreateMkStream(ctx, stream, group, "$")

	// 1. Create a channel to distribute tasks to our internal workers
	taskChan := make(chan redis.XMessage)

	// 2. Start a "Worker Pool" of 3 Goroutines
	var wg sync.WaitGroup
	numWorkers := 3

	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			workerName := fmt.Sprintf("worker-%d", workerID)
			processInternal(ctx, rdb, stream, group, workerName, taskChan)
		}(i)
	}

	// 3. Main Loop: Pull from Redis and push into the internal channel
	log.Printf("[MAIN] Starting dispatcher... Pool size: %d", numWorkers)
	for {
		entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: "dispatcher",
			Streams:  []string{stream, ">"},
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			continue
		}

		for _, streamEntry := range entries {
			for _, message := range streamEntry.Messages {
				// Send the message to the next available internal worker
				taskChan <- message
			}
		}
	}
}

func processInternal(ctx context.Context, rdb *redis.Client, stream, group, name string, ch chan redis.XMessage) {
	for msg := range ch {
		// Extract task data from Redis message
		taskData, ok := msg.Values["task_data"].(string)
		if !ok {
			log.Printf("[%s] Invalid task data in message: %s", name, msg.ID)
			continue
		}

		// Parse task to get the actual task ID
		var t task.Task
		if err := json.Unmarshal([]byte(taskData), &t); err != nil {
			log.Printf("[%s] Failed to parse task: %v", name, err)
			continue
		}

		taskID := t.ID
		log.Printf("[%s] Received Task: %s", name, taskID)

		// 1. Tell the Dashboard we STARTED
		broadcastStatus(ctx, rdb, taskID, "processing", name, "Allocating resources and starting execution")

		// 2. Simulate Work
		time.Sleep(2 * time.Second)
		broadcastStatus(ctx, rdb, taskID, "processing", name, "Analyzing task payload and optimizing")
		time.Sleep(3 * time.Second)

		// 3. Acknowledge in Redis Stream
		rdb.XAck(ctx, stream, group, msg.ID)

		// 4. Tell the Dashboard we FINISHED
		broadcastStatus(ctx, rdb, taskID, "completed", name, "Task execution finalized successfully")
		log.Printf("[%s] Finished & Acked: %s", name, taskID)
	}
}

// Add this helper function to cmd/worker/main.go
func broadcastStatus(ctx context.Context, rdb *redis.Client, taskID, status, worker, message string) {
	update := fmt.Sprintf(`{"id": "%s", "status": "%s", "time": "%s", "worker": "%s", "message": "%s"}`,
		taskID, status, time.Now().Format("15:04:05"), worker, message)

	rdb.Publish(ctx, "task_updates", update)
}
