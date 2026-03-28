package nonblock

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

var (
	ErrNoKeyResponse = errors.New("no key response")

	// ErrNoDataResponse indicates that no more data will be produced.
	// Currently it is aliased to io.EOF for backward compatibility.
	ErrNoDataResponse = io.EOF // errors.New("no data response")
)

type keyResponse struct {
	key string
	err error
}

type dataResponse[T any] struct {
	val T
	err error
}

type NonBlock[T any] struct {
	chKeyReq   chan struct{}
	chKeyRes   chan keyResponse
	chDataRes  chan dataResponse[T]
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	noMoreData bool
}

func New[T any](keyGetter func() (string, error),
	dataGetter func() (T, error)) *NonBlock[T] {

	ctx, cancel := context.WithCancel(context.Background())

	w := &NonBlock[T]{
		chKeyReq:  make(chan struct{}),
		chKeyRes:  make(chan keyResponse),
		chDataRes: make(chan dataResponse[T]),
		cancel:    cancel,
	}
	w.wg.Add(2)

	go func() {
		defer w.wg.Done()
		defer close(w.chKeyRes)
		for {
			select {
			case <-ctx.Done():
				return
			case <-w.chKeyReq:
				key, err := keyGetter()
				select {
				case w.chKeyRes <- keyResponse{key: key, err: err}:
				case <-ctx.Done():
					return
				}

			}
		}
	}()

	go func() {
		defer w.wg.Done()
		defer close(w.chDataRes)
		for {
			if dataGetter == nil {
				return
			}
			data, err := dataGetter()
			select {
			case w.chDataRes <- dataResponse[T]{val: data, err: err}:
			case <-ctx.Done():
				return
			}
			if errors.Is(err, io.EOF) {
				return
			}
		}
	}()

	return w
}

func (w *NonBlock[T]) GetOr(work func(val T, err error) bool) (string, error) {
	w.chKeyReq <- struct{}{}
	if w.noMoreData {
		res, ok := <-w.chKeyRes
		if !ok {
			return "", ErrNoKeyResponse
		}
		return res.key, res.err
	}
	for {
		select {
		case res, ok := <-w.chKeyRes:
			if !ok {
				return "", ErrNoKeyResponse
			}
			return res.key, res.err
		case res, ok := <-w.chDataRes:
			if !ok || work == nil || !work(res.val, res.err) {
				res, ok := <-w.chKeyRes
				w.noMoreData = true
				if !ok {
					return "", ErrNoKeyResponse
				}
				return res.key, res.err
			}
		}
	}
}

func (w *NonBlock[T]) Fetch() (T, error) {
	res, ok := <-w.chDataRes
	if !ok {
		w.noMoreData = true
		var zero T
		return zero, ErrNoDataResponse
	}
	if errors.Is(res.err, io.EOF) {
		w.noMoreData = true
	}
	return res.val, res.err
}

// TryFetch reads a single data item with a timeout.
// This method is intended for use cases where only data retrieval is needed
// and no key input is involved.
// If the timeout expires, it returns os.ErrDeadlineExceeded.
// If the data input channel is closed, it returns io.EOF.
func (w *NonBlock[T]) TryFetch(timeout time.Duration) (T, error) {
	var zero T
	select {
	case res, ok := <-w.chDataRes:
		if !ok {
			w.noMoreData = true
			return zero, ErrNoDataResponse
		}
		if errors.Is(res.err, io.EOF) {
			w.noMoreData = true
		}
		return res.val, res.err
	case <-time.After(timeout):
		return zero, os.ErrDeadlineExceeded
	}
}

func (w *NonBlock[T]) Close() {
	w.cancel()
	w.wg.Wait()
}
