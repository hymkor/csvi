package nonblock

import (
	"errors"
	"io"
	"time"
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
	chStopReq  chan struct{}
	noMoreData bool
}

func New[T any](keyGetter func() (string, error),
	dataGetter func() (T, error)) *NonBlock[T] {

	w := &NonBlock[T]{
		chKeyReq:  make(chan struct{}),
		chKeyRes:  make(chan keyResponse),
		chDataRes: make(chan dataResponse[T]),
		chStopReq: make(chan struct{}),
	}

	go func() {
		for range w.chKeyReq {
			key, err := keyGetter()
			w.chKeyRes <- keyResponse{key: key, err: err}
		}
	}()

	go func() {
		for {
			select {
			case <-w.chStopReq:
				return
			default:
				if dataGetter == nil {
					close(w.chDataRes)
					return
				}
				data, err := dataGetter()
				w.chDataRes <- dataResponse[T]{val: data, err: err}
				if errors.Is(err, io.EOF) {
					close(w.chDataRes)
					return
				}
			}
		}
	}()

	return w
}

func (w *NonBlock[T]) GetOr(work func(val T, err error) bool) (string, error) {
	w.chKeyReq <- struct{}{}
	if w.noMoreData {
		res, ok := <-w.chKeyRes
		if ok {
			return res.key, res.err
		}
		return "", io.EOF
	}
	for {
		select {
		case res, ok := <-w.chKeyRes:
			if ok {
				return res.key, res.err
			}
		case res, ok := <-w.chDataRes:
			if !ok || work == nil || !work(res.val, res.err) {
				res := <-w.chKeyRes
				w.noMoreData = true
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
		return zero, io.EOF
	}
	if errors.Is(res.err, io.EOF) {
		w.noMoreData = true
	}
	return res.val, res.err
}

func (w *NonBlock[T]) TryFetch(timeout time.Duration) (T, error) {
	select {
	case res, ok := <-w.chDataRes:
		if ok {
			if errors.Is(res.err, io.EOF) {
				w.noMoreData = true
			}
			return res.val, res.err
		}
		w.noMoreData = true
	case <-time.After(timeout):
	}
	var zero T
	return zero, io.EOF
}

func (w *NonBlock[T]) Close() {
	close(w.chStopReq)
	close(w.chKeyReq)
}
