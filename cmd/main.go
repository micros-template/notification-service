package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"10.1.20.130/dropping/notification-service/cmd/bootstrap"
	"10.1.20.130/dropping/notification-service/cmd/server"
)

func main() {
	container := bootstrap.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	subscriberReady := make(chan bool)
	subscriberDone := make(chan struct{})
	subscriber := &server.Subscriber{
		Container:       container,
		ConnectionReady: subscriberReady,
	}
	go func() {
		subscriber.Run(ctx)
		close(subscriberDone)
	}()
	<-subscriberReady

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGTERM)
	<-sig

	<-subscriberDone
}
