package server

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitSignal() {
	<-WaitSignalChan()
}

func WaitSignalChan() <-chan os.Signal {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	return stop
}

func GracefulShutdown(name string, shutdown func(ctx context.Context) error) error {
	log.Printf("Shutting down %s server...", name)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := shutdown(ctx); err != nil {
		return err
	}

	log.Printf("%s server stopped gracefully", name)
	return nil
}
