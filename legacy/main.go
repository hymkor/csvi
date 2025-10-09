package legacy

import (
	"github.com/hymkor/csvi"

	"github.com/hymkor/csvi/internal/manualctl"
)

// Terminal is the terminal object not supporting `ESC[6n`
type Terminal struct {
	csvi.Pilot
}

func New() (Terminal, error) {
	p, err := manualctl.New()
	return Terminal{Pilot: p}, err
}

func (L Terminal) Calibrate() error {
	return nil
}
