package main

import (
	"context"
	"events-to-log/events"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var wg sync.WaitGroup

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer stop()

	watcher := events.Init()
	go watcher.StartWatching()

	<-ctx.Done()

	watcher.StopWatching()

	wg.Wait()
	os.Exit(0)
}
