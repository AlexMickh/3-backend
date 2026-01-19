package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexMickh/shop-backend/internal/app"
	"github.com/AlexMickh/shop-backend/internal/config"
	"github.com/AlexMickh/shop-backend/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.New(cfg.Env, os.Stdout)
	log.Info("logger os working")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = logger.ContextWithLogger(ctx, log)

	app := app.New(ctx, cfg)
	app.Run(ctx)
	defer app.GracefulStop(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	close(stop)
	logger.FromCtx(ctx).Info("server stopped")
}
