package candidate

type Candidate []string

func (c Candidate) Len() int {
	return len(c)
}

func (c Candidate) At(n int) string {
	return c[len(c)-n-1]
}

func (c Candidate) Delimiters() string {
	return ""
}

func (c Candidate) Enclosures() string {
	return ""
}

func (c Candidate) List(field []string) (fullnames, basenames []string) {
	return c, c
}
