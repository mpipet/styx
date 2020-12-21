package recioutil

import (
	// "fmt"
	"gitlab.com/dataptive/styx/recio"
)

type Line []byte

func (l *Line) Decode(p []byte) (n int, err error) {

	pos := -1

	for i := 0; i < len(p); i++ {

		if p[i] == 0x0a {
			pos = i
			break
		}

	}

	if pos == -1 {
		return 0, recio.ErrShortBuffer
	}

	*l = Line(p[:pos])

	return pos+1, nil
}

func (l *Line) Encode(p []byte) (n int, err error) {


	if len(*l) + 1 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	n = copy(p, *l)
	p[n] = 0x0a

	n ++

	return n, nil
}
