package csviapp

import (
	"io"
	"strings"

	"github.com/nyaosorg/go-ttyadapter/auto"

	"github.com/hymkor/csvi/candidate"
)

type autoPilot struct {
	*auto.Pilot
}

func newAutoPilot(script string) *autoPilot {
	text := strings.Split(script, "|")
	return &autoPilot{
		Pilot: &auto.Pilot{
			Text: text,
		},
	}
}

func (ap *autoPilot) ReadLine(_ io.Writer, x string, y string, _ candidate.Candidate) (rc string, err error) {
	rc, err = ap.Pilot.GetKey()
	return
}

func (ap *autoPilot) GetFilename(out io.Writer, prompt string, defaultName string) (rc string, err error) {
	rc, err = ap.Pilot.GetKey()
	return
}
