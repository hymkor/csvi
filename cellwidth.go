package csvi

import (
	"strconv"
	"strings"
)

type CellWidth struct {
	Default int
	Option  map[int]int
}

func NewCellWidth() *CellWidth {
	cw := &CellWidth{
		Default: 14,
		Option:  map[int]int{},
	}
	return cw
}

func (cw *CellWidth) Set(at, value int) bool {
	if value == cw.Default {
		delete(cw.Option, at)
	} else {
		cw.Option[at] = value
	}
	return true
}

func (cw *CellWidth) Get(n int) int {
	if val, ok := cw.Option[n]; ok {
		return val
	}
	return cw.Default
}

func (cw *CellWidth) Parse(s string) error {
	var p string
	cont := true
	for cont {
		p, s, cont = strings.Cut(s, ",")
		left, right, ok := strings.Cut(p, ":")
		if ok {
			leftN, err := strconv.ParseUint(left, 10, 64)
			if err != nil {
				return err
			}
			rightN, err := strconv.ParseUint(right, 10, 64)
			if err != nil {
				return err
			}
			cw.Option[int(leftN)] = int(rightN)
		} else {
			value, err := strconv.ParseUint(p, 10, 64)
			if err != nil {
				return err
			}
			cw.Default = int(value)
		}
	}
	return nil
}
