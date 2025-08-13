package task

import (
	"context"
	"sort"
	"sync"
)

type MemRepo struct {
	mu   sync.RWMutex
	seq  int
	data map[int]Task
}

func NewMemRepo() *MemRepo {
	return &MemRepo{
		data: make(map[int]Task),
	}
}

func (r *MemRepo) Create(ctx context.Context, t Task) (Task, error) {
	_ = ctx 

	if err := t.NormalizeAndValidate(); err != nil {
		return Task{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	t.ID = r.seq

	r.data[t.ID] = t
	return t, nil
}

func (r *MemRepo) GetByID(ctx context.Context, id int) (Task, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.data[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (r *MemRepo) List(ctx context.Context, status *Status) ([]Task, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Task, 0, len(r.data))
	if status == nil {
		for _, t := range r.data {
			out = append(out, t)
		}
	} else {
		for _, t := range r.data {
			if t.Status == *status {
				out = append(out, t)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}