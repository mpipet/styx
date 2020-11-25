package recio

import (
	"encoding/binary"
)

// record is a test type that holds a single int that encodes to big endian
// uint32.
type record int

func (r *record) Encode(p []byte) (n int, err error) {

	if len(p) < 4 {
		return 0, ErrShortBuffer
	}

	binary.BigEndian.PutUint32(p[0:4], uint32(*r))

	return 4, nil
}

func (r *record) Decode(p []byte) (n int, err error) {

	if len(p) < 4 {
		return 0, ErrShortBuffer
	}

	*r = record(binary.BigEndian.Uint32(p[0:4]))

	return 4, nil
}

// nullWriter and nullReader are an io.Writer and an io.Reader that always
// succeed. They allow benchmarking the package performance without measuring
// the overhead of an actual writer or reader.
type nullWriter struct{}

func (nw *nullWriter) Write(p []byte) (n int, err error) {

	return len(p), nil
}

type nullReader struct{}

func (nr *nullReader) Read(p []byte) (n int, err error) {

	return len(p), nil
}
