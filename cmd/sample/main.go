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

func WithNotFoundFallback(root http.FileSystem, fallback string) http.Handler {
	f := &customFs{
		fs:       root,
		fallback: fallback,
	}

	return http.FileServer(f)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	contents, err := fs.Sub(client.Content, "build")
	if err != nil {
		log.Fatalf("failed to find assets: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", WithNotFoundFallback(http.FS(contents), "index.html"))
	mux.HandleFunc("/api", api)

	server := &http.Server{
		Addr:    ":9090",
		Handler: mux,
	}

	go func() {
		<-ctx.Done()

		cancelCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(cancelCtx); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}
