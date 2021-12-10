//go:build !windows
// +build !windows

package main

import (
	"strings"
)

func textfilter(s string) (string, _CodeFlag) {
	if strings.HasPrefix(s, bomCode) {
		return s[len(bomCode):], hasBom
	}
	return s, nonBomUtf8
}
