package queue

import (
	"context"
	"encoding/json"

	"github.com/heyu255/distritask/internal/task"
	"github.com/redis/go-redis/v9"
)

type RedisQueue struct {
	client *redis.Client
	stream string
}

func NewRedisQueue(client *redis.Client, stream string) *RedisQueue {
	return &RedisQueue{
		client: client,
		stream: stream,
	}
}

func (rq *RedisQueue) Enqueue(ctx context.Context, t *task.Task) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return rq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rq.stream,
		Values: map[string]interface{}{"task_data": data},
	}).Err()
}
