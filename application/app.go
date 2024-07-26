package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		rdb: redis.NewClient(&redis.Options{}),
	}
	app.loadroutes()
	return app
}

func (a *App) Start(ctx context.Context) error {
	S := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := a.rdb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %w", err)
	}

	defer func() {
		if err := a.rdb.Close(); err != nil {
			fmt.Println("failed to close redis", err)
		}

	}()
	fmt.Println("Server started")

	ch := make(chan error, 1)
	go func() {
		err = S.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to connect %w", err)
		}
		close(ch)

	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return S.Shutdown(timeout)
	}
}
