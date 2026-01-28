package manualctl

import (
	"github.com/nyaosorg/go-ttyadapter"
)

// pendingEscTty buffers the ESC key to avoid misinterpreting
// partially received key sequences.
type pendingEscTty struct {
	ttyadapter.Tty
}

// GetKey returns only when a complete key sequence is received.
// Partial escape sequences are buffered internally.
func (p pendingEscTty) GetKey() (string, error) {
	var sequence string
	for {
		key, err := p.Tty.GetKey()
		if err != nil {
			return "", err
		}
		sequence += key
		switch sequence {
		case "\x1B", "\x1B[", "\x1B[1", "\x1B[15", "\x1B[16", "\x1B[17",
			"\x1B[18", "\x1B[1;", "\x1B[1;5", "\x1B[2", "\x1B[20",
			"\x1B[21", "\x1B[23", "\x1B[24", "\x1B[3", "\x1B[5",
			"\x1B[5;", "\x1B[5;5", "\x1B[6", "\x1B[6;", "\x1B[6;5", "\x1B[O":
		default:
			return sequence, nil
		}
	}
}
