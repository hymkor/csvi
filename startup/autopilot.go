package startup

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/hymkor/csvi/candidate"
)

type autoPilot struct {
	script string
}

func (ap *autoPilot) Size() (int, int, error) {
	return 80, 25, nil
}

func (ap *autoPilot) next() (string, error) {
	if ap.script == "" {
		return "", io.EOF
	}
	var command string
	command, ap.script, _ = strings.Cut(ap.script, "|")
	return command, nil
}

func (ap *autoPilot) ReadLine(io.Writer, string, string, candidate.Candidate) (string, error) {
	return ap.next()
}

func (ap *autoPilot) GetKey() (string, error) {
	key, err := ap.next()
	if err != nil || len(key) <= 1 || key[0] == '\x1B' {
		return key, err
	}
	if utf8.RuneCountInString(key) != 1 {
		return key, fmt.Errorf("%#v: too long string for getkey", key)
	}
	return key, nil
}

func (ap *autoPilot) GetFilename(out io.Writer, prompt string, defaultName string) (string, error) {
	return ap.next()
}

func (ap *autoPilot) Close() error {
	return nil
}
