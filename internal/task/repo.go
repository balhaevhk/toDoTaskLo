package task

// internal/task/repo.go
type Repository interface {
    Create(ctx context.Context, t Task) (Task, error)
    GetByID(ctx context.Context, id int) (Task, error)
    List(ctx context.Context, status *string) ([]Task, error)
}

// internal/task/memrepo.go
func NewMemRepo() Repository { /* ... */ }