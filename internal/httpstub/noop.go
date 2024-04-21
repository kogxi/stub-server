package httpstub

import "io"

type Writer struct{}

func newWriter() io.Writer {
	return &Writer{}
}

func (w *Writer) Write(b []byte) (int, error) {
	return len(b), nil
}
