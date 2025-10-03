package application

import (
	"net/http"
	"fmt"
	"context"
)

type App struct {
	Router http.Handler
}

func New() *App {
	app := &App{
		Router: loadRoutes(),
	}
	return app
}

func(a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.Router,
	}
	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}