package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"lo/internal/httpapi"
	"lo/internal/logasync"
	"lo/internal/task"
)

func main() {
	repo := task.NewMemRepo()
	logger := logasync.New(256)

	h := httpapi.NewHandler(repo, logger)
	mux := httpapi.NewMux(h)

	srv := &http.Server{
		Addr:         addr(),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown: %v", err)
	}

	logger.Close()
	log.Printf("server exited")
}

func addr() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}