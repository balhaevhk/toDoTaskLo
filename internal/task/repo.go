package task

import (
	"context"
	"errors"
)

type Repository interface {
	Create(ctx context.Context, t Task) (Task, error)
	GetByID(ctx context.Context, id int) (Task, error)
	List(ctx context.Context, status *Status) ([]Task, error)
}

var ErrNotFound = errors.New("task not found")