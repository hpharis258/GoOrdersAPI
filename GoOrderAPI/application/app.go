package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	Router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		Router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{Addr: "localhost:6379"}), // Example Redis client
	}
	return app
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.Router,
	}
	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	defer func() {
		if cerr := a.rdb.Close(); cerr != nil {
			fmt.Println("Error closing Redis client:", cerr)
		}
	}()

	fmt.Println("Server starting")

	ch := make(chan error, 1)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(timeout)

	}

	return nil
}
