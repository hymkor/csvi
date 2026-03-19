package csvi

import (
	"context"
	"os"
	"os/signal"
	"sync"
)

type ScopedInterrupt struct {
	ch chan os.Signal
}

func NewScopedInterrupt() *ScopedInterrupt {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	return &ScopedInterrupt{
		ch: ch,
	}
}

func (si *ScopedInterrupt) Close() {
	signal.Stop(si.ch)
}

func drainSignal(ch <-chan os.Signal) {
	for {
		select {
		case <-ch:
			// drop
		default:
			return
		}
	}
}

func (si *ScopedInterrupt) NotifyContext(ctx context.Context) (context.Context, context.CancelFunc) {
	drainSignal(si.ch)

	_ctx, cancel := context.WithCancel(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case <-si.ch:
			cancel()
		case <-_ctx.Done():
		}
	}()

	return _ctx, func() {
		cancel()
		wg.Wait()
	}
}
