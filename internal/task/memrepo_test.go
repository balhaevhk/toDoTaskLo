package task_test

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"lo/internal/task"
)

func TestCreate_AutoIncrementAndDefaults(t *testing.T) {
	t.Parallel()

	repo := task.NewMemRepo()

	ctx := context.Background()
	input := task.Task{
		Title:       "First",
		Description: "desc",
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected non-zero auto-increment ID, got %d", created.ID)
	}
	if created.Title != input.Title {
		t.Errorf("Title mismatch: want %q, got %q", input.Title, created.Title)
	}
	if created.Description != input.Description {
		t.Errorf("Description mismatch: want %q, got %q", input.Description, created.Description)
	}
	if created.Status == "" {
		t.Errorf("expected default status to be set, got empty")
	}
	if time.Since(created.CreatedAt) > time.Second {
		t.Errorf("CreatedAt seems off: %v", created.CreatedAt)
	}
}

func TestGetByID_FoundAndNotFound(t *testing.T) {
	t.Parallel()

	repo := task.NewMemRepo()
	ctx := context.Background()

	one, err := repo.Create(ctx, task.Task{Title: "A"})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}

	got, err := repo.GetByID(ctx, one.ID)
	if err != nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got.ID != one.ID {
		t.Errorf("GetByID mismatch: want ID %d, got %d", one.ID, got.ID)
	}

	_, err = repo.GetByID(ctx, 9999)
	if err == nil {
		t.Fatalf("expected error for not found id")
	}
	var notFoundErr interface{ NotFound() bool }
	if errors.As(err, &notFoundErr) && !notFoundErr.NotFound() {
		t.Errorf("expected NotFound error, got: %v", err)
	}
}

func TestList_FilterByStatus(t *testing.T) {
	t.Parallel()

	repo := task.NewMemRepo()
	ctx := context.Background()

	mustCreate := func(title string, status task.Status) int {
		t.Helper()
		created, err := repo.Create(ctx, task.Task{Title: title, Status: status})
		if err != nil {
			t.Fatalf("Create error: %v", err)
		}
		return created.ID
	}

	idNew1 := mustCreate("n1", task.StatusNew)
	_ = mustCreate("ip1", task.StatusInProgress)
	idNew2 := mustCreate("n2", task.StatusNew)
	_ = mustCreate("done1", task.StatusDone)

	all, err := repo.List(ctx, nil)
	if err != nil {
		t.Fatalf("List all error: %v", err)
	}
	if len(all) != 4 {
		t.Fatalf("want 4 tasks, got %d", len(all))
	}

	newStatus := task.StatusNew
	onlyNew, err := repo.List(ctx, &newStatus)
	if err != nil {
		t.Fatalf("List filtered error: %v", err)
	}
	gotIDs := make([]int, 0, len(onlyNew))
	for _, tsk := range onlyNew {
		if tsk.Status != task.StatusNew {
			t.Fatalf("unexpected status in result: %q", tsk.Status)
		}
		gotIDs = append(gotIDs, tsk.ID)
	}
	sort.Ints(gotIDs)
	want := []int{idNew1, idNew2}
	if len(gotIDs) != len(want) || gotIDs[0] != want[0] || gotIDs[1] != want[1] {
		t.Errorf("filtered IDs mismatch: want %v, got %v", want, gotIDs)
	}
}

func TestCreate_ConcurrentUniqueIDs(t *testing.T) {
	t.Parallel()

	repo := task.NewMemRepo()
	ctx := context.Background()

	const n = 100
	errCh := make(chan error, n)
	ids := make(chan int, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			created, err := repo.Create(ctx, task.Task{Title: "t"})
			if err != nil {
				errCh <- err
				return
			}
			ids <- created.ID
			errCh <- nil
		}(i)
	}

	for i := 0; i < n; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("Create concurrent error: %v", err)
		}
	}

	close(ids)
	seen := make(map[int]struct{}, n)
	for id := range ids {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate ID detected: %d", id)
		}
		seen[id] = struct{}{}
	}
}