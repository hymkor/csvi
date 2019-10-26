package main

import (
	"strings"

	"github.com/mattn/go-tty"
)

func getKey(tty1 *tty.TTY) (string, error) {
	clean, err := tty1.Raw()
	if err != nil {
		return "", err
	}
	defer clean()

	var buffer strings.Builder
	escape := false
	for {
		r, err := tty1.ReadRune()
		if err != nil {
			return "", err
		}
		if r == 0 {
			continue
		}
		buffer.WriteRune(r)
		if r == '\x1B' {
			escape = true
		}
		if !(escape && tty1.Buffered()) && buffer.Len() > 0 {
			return buffer.String(), nil
		}
	}
}

func ignoreSigwinch(tty1 *tty.TTY) func() {
	quit := make(chan struct{})
	ws := tty1.SIGWINCH()
	go func() {
		for {
			select {
			case <-quit:
				return
			case <-ws:
			}
		}
	}()

	return func() {
		quit <- struct{}{}
	}
}
