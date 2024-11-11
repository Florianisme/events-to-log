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

	wg.Add(1)

	watcher := events.Init()
	go watcher.StartWatching(&wg)

	<-ctx.Done()

	watcher.StopWatching()

	wg.Wait()
}
