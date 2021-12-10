package main

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/sys/windows"
)

func atou(bin []byte) (string, error) {
	acp := windows.GetACP()
	length, err := windows.MultiByteToWideChar(acp, 0, &bin[0], int32(len(bin)), nil, 0)
	if err != nil {
		return "", err
	}
	buffer := make([]uint16, length)
	_, err = windows.MultiByteToWideChar(acp, 0, &bin[0], int32(len(bin)), &buffer[0], length)
	if err != nil {
		return "", err
	}
	return windows.UTF16ToString(buffer), nil
}

func textfilter(s string) (string, _CodeFlag) {
	codeFlag := nonBomUtf8
	if utf8.ValidString(s) {
		if strings.HasPrefix(s, bomCode) {
			s = s[len(bomCode):]
			codeFlag |= hasBom
		}
	} else if ansiStr, err := atou([]byte(s)); err == nil {
		s = ansiStr
		codeFlag |= isAnsi
	}
	return s, codeFlag
}
