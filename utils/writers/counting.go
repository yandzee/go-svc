package writerutils

import (
	"io"
	"strconv"
)

type CountingWriter struct {
	io.Writer
	nbytes int
}

func (cw *CountingWriter) Write(chunk []byte) (int, error) {
	n, err := cw.Writer.Write(chunk)
	cw.nbytes += n

	return n, err
}

func (cw *CountingWriter) Count() int {
	return cw.nbytes
}

func (cw *CountingWriter) CountString() string {
	return strconv.Itoa(cw.nbytes)
}
