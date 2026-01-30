package csvi

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

const (
	voicedSoundMark     = "\uFF9E"
	semiVoicedSoundMark = "\uFF9F"
	delChar             = "\x7F"
)

func insDel(s string, mark string) string {
	var buffer strings.Builder
	for {
		at := strings.Index(s, mark)
		if at < 0 {
			if buffer.Len() > 0 {
				buffer.WriteString(s)
				return buffer.String()
			}
			return s
		}
		buffer.WriteString(s[:at])
		buffer.WriteString(delChar)
		buffer.WriteString(mark)
		s = s[at+len(mark):]
	}
}

// Temporarily insert DEL (U+007F) before halfwidth voiced/semi-voiced marks
// to prevent runewidth from treating them as emoji-like ligatures.
// DEL is chosen because it has zero width and no terminal side effects.
func truncate(s string, w int, tail string) string {
	s = insDel(s, voicedSoundMark)
	s = insDel(s, semiVoicedSoundMark)
	s = runewidth.Truncate(s, w, tail)
	return s
}
