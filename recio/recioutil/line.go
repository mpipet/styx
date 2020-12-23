package recioutil

import (
	"gitlab.com/dataptive/styx/recio"
)

var (
	LineEndCR = []byte{0x0d}
	LineEndLF = []byte{0x0a}
	LineEndCRLF = []byte{0x0d, 0x0a}
)

type LineCR []byte

func (l *LineCR) Decode(p []byte) (n int, err error) {

	pos := -1

	for i := 0; i < len(p); i++ {

		if p[i] == LineEndCR[0] {
			pos = i
			break
		}

	}

	if pos == -1 {
		return 0, recio.ErrShortBuffer
	}

	*l = LineCR(p[:pos])

	return pos+1, nil
}

func (l *LineCR) Encode(p []byte) (n int, err error) {

	if len(*l) + 1 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	n = copy(p, *l)
	p[n] = LineEndCR[0]

	n ++

	return n, nil
}

type LineLF []byte

func (l *LineLF) Decode(p []byte) (n int, err error) {

	pos := -1

	for i := 0; i < len(p); i++ {

		if p[i] == LineEndLF[0] {
			pos = i
			break
		}

	}

	if pos == -1 {
		return 0, recio.ErrShortBuffer
	}

	*l = LineLF(p[:pos])

	return pos+1, nil
}

func (l *LineLF) Encode(p []byte) (n int, err error) {

	if len(*l) + 1 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	n = copy(p, *l)
	p[n] = LineEndLF[0]

	n ++

	return n, nil
}

type LineCRLF []byte

func (l *LineCRLF) Decode(p []byte) (n int, err error) {

	pos := -1
	delimPos := 0

	for i := 0; i < len(p); i++ {

		if p[i] == LineEndCRLF[delimPos] {
			if delimPos == 0 {
				pos = i
			}

			delimPos++
		} else {
			delimPos = 0
			pos = -1
		}

		if delimPos == 2 {
			break
		}
	}

	if delimPos != 2 {
		return 0, recio.ErrShortBuffer
	}

	*l = LineCRLF(p[:pos])

	return pos+2, nil
}

func (l *LineCRLF) Encode(p []byte) (n int, err error) {

	if len(*l) + 2 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	n = copy(p, *l)
	p[n] = LineEndCRLF[0]
	n ++
	p[n] = LineEndCRLF[1]
	n ++

	return n, nil
}
