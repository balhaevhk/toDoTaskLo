package logasync

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Logger struct {
	ch   chan Event
	wg   sync.WaitGroup
	once sync.Once
}

func New(size int) *Logger {
	l := &Logger{ch: make(chan Event, size)}
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		enc := json.NewEncoder(os.Stdout)
		for e := range l.ch {
			if e.TS.IsZero() {
				e.TS = time.Now()
			}
			_ = enc.Encode(e)
		}
	}()
	return l
}

func (l *Logger) Publish(e Event) {
	select {
		case l.ch <- e:
		default:
	}
}

func (l *Logger) Close() {
	l.once.Do(func() {
		close(l.ch)
		l.wg.Wait()
	})
}