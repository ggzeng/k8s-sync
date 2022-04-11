package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"k8sync/internal/config"
	"k8sync/internal/gateway"
	"k8sync/internal/k8s/client"
	"k8sync/internal/k8s/controller"
	"k8sync/internal/k8s/handler"
	log "k8sync/pkg/logger"
)

func daemonStart(cmd *cobra.Command, args []string) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(2 * time.Second)
		log.Info("daemon stopped")
	}()

	if err = gateway.Start(ctx); err != nil {
		log.Error(err)
		return
	}
	log.Info("grpc svc started")

	var handler handler.Handler
	if handler, err = prepareHandler(); err != nil {
		log.Error(err)
		return
	}
	defer handler.Clean()

	k8s := client.New("src")
	controller.Start(ctx, k8s, handler)
	log.Info("k8s controller started")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-done
}

func prepareHandler() (handler.Handler, error) {
	handler := new(handler.Default)
	if err := handler.Init(config.Curr()); err != nil {
		return nil, fmt.Errorf("init handler failed: %w", err)
	}
	return handler, nil
}
