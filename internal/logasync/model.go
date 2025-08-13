package logasync

import "time"

type Level string

const (
	LevelInfo  Level = "info"
	LevelError Level = "error"
)

type Event struct {
	TS         time.Time `json:"ts"`
	Level      Level     `json:"level"`
	Action     string    `json:"action,omitempty"`
	Method     string    `json:"method,omitempty"`
	Path       string    `json:"path,omitempty"`
	TaskID     *int      `json:"task_id,omitempty"`
	Status     string    `json:"status,omitempty"`
	HTTPStatus int       `json:"http_status,omitempty"`
	LatencyMS  int64     `json:"latency_ms,omitempty"`
	Err        string    `json:"err,omitempty"`
}