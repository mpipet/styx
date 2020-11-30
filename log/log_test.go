package log

import (
	"os"
	"path/filepath"
	"testing"

	"gitlab.com/dataptive/styx/recio"
)

// Benchmarks auto and unsafe writes of records of varying sizes to a
// LogWriter.
func BenchmarkLog_WriterUnsafe10(b *testing.B) {
	benchmarkLog_Writer(b, 10, SyncUnsafe)
}

func BenchmarkLog_WriterUnsafe100(b *testing.B) {
	benchmarkLog_Writer(b, 100, SyncUnsafe)
}

func BenchmarkLog_WriterUnsafe500(b *testing.B) {
	benchmarkLog_Writer(b, 500, SyncUnsafe)
}

func BenchmarkLog_WriterUnsafe1000(b *testing.B) {
	benchmarkLog_Writer(b, 1000, SyncUnsafe)
}

func BenchmarkLog_WriterAuto10(b *testing.B) {
	benchmarkLog_Writer(b, 10, SyncAuto)
}

func BenchmarkLog_WriterAuto100(b *testing.B) {
	benchmarkLog_Writer(b, 100, SyncAuto)
}

func BenchmarkLog_WriterAuto500(b *testing.B) {
	benchmarkLog_Writer(b, 500, SyncAuto)
}

func BenchmarkLog_WriterAuto1000(b *testing.B) {
	benchmarkLog_Writer(b, 1000, SyncAuto)
}

func benchmarkLog_Writer(b *testing.B, payloadSize int, syncMode SyncMode) {

	b.StopTimer()

	// XXX: b.TempDir() fails when doing multiple benchmarks on current
	// go version (1.15.4).
	path := "tmp"
	err := os.Mkdir(path, os.FileMode(0744))
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)

	name := filepath.Join(path, "bench")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1 << 20, syncMode, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	b.StartTimer()

	written := int64(0)
	for i := 0; i < b.N; i++ {
		n, err := lw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}

		written += int64(n)
	}

	err = lw.Flush()
	if err != nil {
		b.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(written / int64(b.N))
}

// Benchmarks reads of records of varying sizes from a LogReader.
func BenchmarkLog_Reader10(b *testing.B) {
	benchmarkLog_Reader(b, 10)
}

func BenchmarkLog_Reader100(b *testing.B) {
	benchmarkLog_Reader(b, 100)
}

func BenchmarkLog_Reader500(b *testing.B) {
	benchmarkLog_Reader(b, 500)
}

func BenchmarkLog_Reader1000(b *testing.B) {
	benchmarkLog_Reader(b, 1000)
}

func benchmarkLog_Reader(b *testing.B, payloadSize int) {

	b.StopTimer()

	// XXX: b.TempDir() fails when doing multiple benchmarks on current
	// go version (1.15.4).
	path := "tmp"
	err := os.Mkdir(path, os.FileMode(0744))
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)

	name := filepath.Join(path, "bench")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1 << 20, SyncManual, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	for i := 0; i < b.N; i++ {
		_, err := lw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}
	}

	err = lw.Flush()
	if err != nil {
		b.Fatal(err)
	}

	err = lw.Sync()
	if err != nil {
		b.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		b.Fatal(err)
	}

	lr, err := l.NewReader(1 << 20, true, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	read := int64(0)
	for i := 0; i < b.N; i++ {
		n, err := lr.Read(&r)
		if err != nil {
			b.Fatal(err)
		}

		read += int64(n)
	}

	err = lr.Close()
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(read / int64(b.N))
}
