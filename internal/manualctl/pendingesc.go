package manualctl

import (
	"github.com/nyaosorg/go-ttyadapter"
)

// PendingEscTty buffers the ESC key to avoid misinterpreting
// partially received key sequences.
type PendingEscTty struct {
	ttyadapter.Tty
	pending string
}

func (p *PendingEscTty) GetKey() (string, error) {
	key, err := p.Tty.GetKey()
	join := p.pending + key
	switch join {
	case "\x1B", "\x1B[":
		p.pending = join
		key = ""
	default:
		p.pending = ""
		key = join
	}
	return key, err
}
