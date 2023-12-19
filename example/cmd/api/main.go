package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/smiletrl/mq"
	"github.com/smiletrl/mq/example/pkg/logger"
	"github.com/smiletrl/mq/example/pkg/postgres"
	notify "github.com/smiletrl/mq/example/service.notify"
	order "github.com/smiletrl/mq/example/service.order"
)

func main() {
	// pgx db pool
	pool, err := postgres.NewPool()
	if err != nil {
		panic("pgx pool instance error: " + err.Error())
	}

	logger := logger.NewLogger()

	mqProvider := mq.NewProvider(pool)

	// echo instance
	e := echo.New()
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// health check
	e.GET("", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	group := e.Group("/api")

	// register handlers
	orderRepo := order.NewRepository(pool)
	orderSvc := order.NewService(pool, orderRepo, mqProvider)
	order.RegisterHandlers(group, orderSvc)

	// register consumers
	notifySvc := notify.NewService()
	notify.RegisterConsumer(notifySvc)

	order.RegisterConsumer(orderRepo)

	// Start rest server
	go func() {
		err := e.Start(":1325")
		if err != nil {
			log.Error("failed to start echo", err)
		}
	}()

	// start mq consumer
	go func() {
		consume := mq.NeWConsume(pool, logger)
		consume.Consume()
	}()

	// gracefully shutdown application
	shutdown(e)
}

func shutdown(e *echo.Echo) {
	// Handle SIGTERM
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ch:
		e.Shutdown(context.Background())
	}
}
