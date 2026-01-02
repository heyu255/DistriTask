package task

import "time"

type State int

const (
	Pending State = iota
	Running
	Completed
	Failed
)

type Task struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Payload     []byte    `json:"payload"`
	Status      State     `json:"status"`
	Retries     int       `json:"retries"`
	MaxRetries  int       `json:"max_retries"`
	CreatedAt   time.Time `json:"created_at"`
}