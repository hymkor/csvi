package csvi

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

const (
	voicedSoundMark     = "\uFF9E"
	semiVoicedSoundMark = "\uFF9F"
	privateUse1         = "\uE000"
	privateUse2         = "\uE001"
)

var rewindTable = strings.NewReplacer(
	privateUse1, voicedSoundMark,
	privateUse2, semiVoicedSoundMark,
)

// Temporarily replace halfwidth voiced/semi-voiced sound marks
// with private-use characters to avoid width underestimation
// in runewidth.Truncate.
func truncate(s string, w int, tail string) string {
	rewind := false
	for {
		at := strings.Index(s, voicedSoundMark)
		if at < 0 {
			break
		}
		s = s[:at] + privateUse1 + s[at+len(voicedSoundMark):]
		rewind = true
	}
	for {
		at := strings.Index(s, semiVoicedSoundMark)
		if at < 0 {
			break
		}
		s = s[:at] + privateUse2 + s[at+len(semiVoicedSoundMark):]
		rewind = true
	}
	s = runewidth.Truncate(s, w, tail)
	if rewind {
		s = rewindTable.Replace(s)
	}
	return s
}
