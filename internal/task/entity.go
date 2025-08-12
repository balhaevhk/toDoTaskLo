package task

// internal/task/entity.go
type Task struct {
    ID          int
    Title       string
    Description string
    Status      string // "new" | "in_progress" | "done"
    CreatedAt   time.Time
}