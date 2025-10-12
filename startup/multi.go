package startup

import (
	"io"
	"os"
)

type stream struct {
	fname string
	fd    *os.File
}

func (s *stream) Read(r []byte) (int, error) {
	if s.fd == nil {
		var err error
		s.fd, err = os.Open(s.fname)
		if err != nil {
			return 0, err
		}
	}
	n, err := s.fd.Read(r)
	if err != nil {
		s.fd.Close()
	}
	return n, err
}

func multiFileReader(filenames ...string) io.Reader {
	if len(filenames) <= 0 {
		return os.Stdin
	}
	ss := make([]io.Reader, 0, len(filenames))
	for _, fn := range filenames {
		ss = append(ss, &stream{fname: fn})
	}
	return io.MultiReader(ss...)
}
