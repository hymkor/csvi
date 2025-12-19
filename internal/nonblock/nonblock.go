package nonblock

import (
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
	chKeyReq  chan struct{}
	chKeyRes  chan keyResponse
	chDataRes chan dataResponse[T]
	chStopReq chan struct{}
}

func New[T any](keyGetter func() (string, error),
	dataGetter func() (bool, T, error)) *NonBlock[T] {

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
				ok, data, err := dataGetter()
				if !ok {
					close(w.chDataRes)
					return
				}
				w.chDataRes <- dataResponse[T]{val: data, err: err}
			}
		}
	}()

	return w
}

func (w *NonBlock[T]) GetOr(work func(val T, err error) bool) (string, error) {
	w.chKeyReq <- struct{}{}
	for {
		select {
		case res, ok := <-w.chKeyRes:
			if ok {
				return res.key, res.err
			}
		case res, ok := <-w.chDataRes:
			if ok && !work(res.val, res.err) {
				res := <-w.chKeyRes
				return res.key, res.err
			}
		}
	}
}

func (w *NonBlock[T]) Fetch() (bool, T, error) {
	res, ok := <-w.chDataRes
	return ok, res.val, res.err
}

func (w *NonBlock[T]) TryFetch(timeout time.Duration) (bool, T, error) {
	select {
	case res, ok := <-w.chDataRes:
		return ok, res.val, res.err
	case <-time.After(timeout):
		var zero T
		return false, zero, nil
	}
}

func (w *NonBlock[T]) Close() {
	close(w.chStopReq)
	close(w.chKeyReq)
}
