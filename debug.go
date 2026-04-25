//go:build debug

package csvi

import (
	"github.com/nyaosorg/go-windows-dbg"
)

func debug(v ...any) {
	dbg.Println(v...)
}
