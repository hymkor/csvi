//go:build !windows
// +build !windows

package main

func textfilter(s string) (string, bool) {
	return s, false
}
