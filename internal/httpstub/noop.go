package httpstub

import "io"

// Writer is a no-op writer that discards all data written to it.
type Writer struct{}

func newWriter() io.Writer {
	return &Writer{}
}

func (w *Writer) Write(b []byte) (int, error) {
	return len(b), nil
}
