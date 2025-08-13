package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"lo/internal/logasync"
	"lo/internal/task"
)

type Handler struct {
	Repo   task.Repository
	Logger *logasync.Logger
}

func NewHandler(repo task.Repository, logger *logasync.Logger) *Handler {
	return &Handler{Repo: repo, Logger: logger}
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	qStatus := r.URL.Query().Get("status")
	var stPtr *task.Status
	if qStatus != "" {
		st := task.Status(qStatus)
		stPtr = &st
	}

	items, err := h.Repo.List(r.Context(), stPtr)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, "list failed", err)
		h.log(r.Context(), logasync.LevelError, "list", r, http.StatusInternalServerError, start, nil, qStatus, err)
		return
	}
	h.writeJSON(w, http.StatusOK, items)
	h.log(r.Context(), logasync.LevelInfo, "list", r, http.StatusOK, start, nil, qStatus, nil)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	id, ok := parseID(r.URL.Path, "/tasks/")
	if !ok {
		h.writeError(w, r, http.StatusNotFound, "not found", nil)
		h.log(r.Context(), logasync.LevelError, "get", r, http.StatusNotFound, start, nil, "", errors.New("route not matched"))
		return
	}

	tk, err := h.Repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, task.ErrNotFound) {
			h.writeError(w, r, http.StatusNotFound, "task not found", nil)
			h.log(r.Context(), logasync.LevelInfo, "get", r, http.StatusNotFound, start, &id, "", err)
			return
		}
		h.writeError(w, r, http.StatusInternalServerError, "get failed", err)
		h.log(r.Context(), logasync.LevelError, "get", r, http.StatusInternalServerError, start, &id, "", err)
		return
	}

	h.writeJSON(w, http.StatusOK, tk)
	h.log(r.Context(), logasync.LevelInfo, "get", r, http.StatusOK, start, &id, string(tk.Status), nil)
}

type createTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status,omitempty"`
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	if r.Method != http.MethodPost {
		h.writeError(w, r, http.StatusMethodNotAllowed, "method not allowed", nil)
		h.log(r.Context(), logasync.LevelError, "create", r, http.StatusMethodNotAllowed, start, nil, "", errors.New("method not allowed"))
		return
	}

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, r, http.StatusBadRequest, "invalid json", err)
		h.log(r.Context(), logasync.LevelError, "create", r, http.StatusBadRequest, start, nil, "", err)
		return
	}
	defer r.Body.Close()

	tk := task.Task{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Status:      task.Status(strings.TrimSpace(req.Status)),
	}
	if tk.Title == "" {
		h.writeError(w, r, http.StatusBadRequest, "title is required", nil)
		h.log(r.Context(), logasync.LevelError, "create", r, http.StatusBadRequest, start, nil, string(tk.Status), errors.New("title required"))
		return
	}

	created, err := h.Repo.Create(r.Context(), tk)
	if err != nil {
		if errors.Is(err, task.ErrInvalidStatus) {
			h.writeError(w, r, http.StatusBadRequest, "invalid status", err)
			h.log(r.Context(), logasync.LevelError, "create", r, http.StatusBadRequest, start, nil, string(tk.Status), err)
			return
		}
		h.writeError(w, r, http.StatusInternalServerError, "create failed", err)
		h.log(r.Context(), logasync.LevelError, "create", r, http.StatusInternalServerError, start, nil, string(tk.Status), err)
		return
	}

	w.Header().Set("Location", "/tasks/"+strconv.Itoa(created.ID))
	h.writeJSON(w, http.StatusCreated, created)
	id := created.ID
	h.log(r.Context(), logasync.LevelInfo, "create", r, http.StatusCreated, start, &id, string(created.Status), nil)
}


func (h *Handler) writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

type errResp struct {
	Error string `json:"error"`
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, code int, msg string, cause error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errResp{Error: msg})
}

func (h *Handler) log(ctx context.Context, lvl logasync.Level, action string, r *http.Request, httpCode int, start time.Time, taskID *int, statusStr string, cause error) {
	var errStr string
	if cause != nil {
		errStr = cause.Error()
	}
	h.Logger.Publish(logasync.Event{
		Level:      lvl,
		Action:     action,
		Method:     r.Method,
		Path:       r.URL.Path,
		TaskID:     taskID,
		Status:     statusStr,
		HTTPStatus: httpCode,
		LatencyMS:  time.Since(start).Milliseconds(),
		Err:        errStr,
	})
}

func parseID(path, prefix string) (int, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	idStr := strings.TrimPrefix(path, prefix)
	if idStr == "" || strings.Contains(idStr, "/") {
		return 0, false
	}
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}