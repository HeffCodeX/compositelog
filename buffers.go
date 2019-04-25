package compositelog

import (
	"bytes"
	"io"
)

type buffers struct {
	ErrorWriters     []io.Writer
	InfoWriters      []io.Writer
	CompositeWriters []io.Writer

	errorBuffer     bytes.Buffer
	infoBuffer      bytes.Buffer
	compositeBuffer bytes.Buffer
}

func (b *buffers) writeError(err error, addNewLine bool) {
	strError := err.Error()
	if addNewLine {
		strError += NewLine
	}

	b.errorBuffer.Write([]byte(strError))
	b.writeComposite(strError, false)
}

func (b *buffers) writeInfo(msg string, addNewLine bool) {
	if addNewLine {
		msg += NewLine
	}

	b.errorBuffer.Write([]byte(msg))
	b.writeComposite(msg, false)
}

func (b *buffers) writeComposite(msg string, addNewLine bool) {
	if addNewLine {
		msg += NewLine
	}

	b.compositeBuffer.Write([]byte(msg))
}

func (b *buffers) flush() { // TODO: async write
	for _, w := range b.ErrorWriters {
		_, _ = w.Write(b.errorBuffer.Bytes())
	}
	b.errorBuffer.Reset()

	for _, w := range b.InfoWriters {
		_, _ = w.Write(b.infoBuffer.Bytes())
	}
	b.infoBuffer.Reset()

	for _, w := range b.CompositeWriters {
		_, _ = w.Write(b.compositeBuffer.Bytes())
	}
	b.compositeBuffer.Reset()
}
