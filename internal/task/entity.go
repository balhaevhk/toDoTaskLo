package task

import (
	"errors"
	"time"
)

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

var validStatuses = map[Status]struct{}{
	StatusNew:        {},
	StatusInProgress: {},
	StatusDone:       {},
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

var ErrInvalidStatus = errors.New("invalid status")

func (t *Task) NormalizeAndValidate() error {
	if t.Status == "" {
		t.Status = StatusNew
	}
	if _, ok := validStatuses[t.Status]; !ok {
		return ErrInvalidStatus
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	return nil
}