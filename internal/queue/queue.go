package queue

import (
	"context"

	"github.com/heyu255/distritask/internal/task"
)

type Queue interface {
	Enqueue(ctx context.Context, t *task.Task) error
}
