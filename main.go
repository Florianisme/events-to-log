package main

import (
	"context"
	"events-to-log/events"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	watcher := events.Init()
	go watcher.StartWatching()

	<-ctx.Done()

	watcher.StopWatching()

	wg.Wait()
}
