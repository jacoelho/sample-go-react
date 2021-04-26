package main

import (
	"context"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jacoelho/sample/client"
)

func api(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("api endpoint"))
	w.WriteHeader(http.StatusOK)
}

type customFs struct {
	fs       http.FileSystem
	fallback string
}

func (c *customFs) Open(name string) (http.File, error) {
	f, err := c.fs.Open(name)

	if errors.Is(err, os.ErrNotExist) {
		return c.fs.Open(c.fallback)
	}

	return f, err
}

func WithNotFoundFallbackFileSystem(root http.FileSystem, fallback string) http.FileSystem {
	return &customFs{
		fs:       root,
		fallback: fallback,
	}
}

func FileServer(contents fs.FS) http.Handler {
	return http.FileServer(WithNotFoundFallbackFileSystem(http.FS(contents), "index.html"))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	contents, err := client.Contents()
	if err != nil {
		log.Fatalf("failed to find assets: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", FileServer(contents))
	mux.HandleFunc("/api", api)

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
