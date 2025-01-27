package utils

import (
	"bytes"
	"io"
)

// Prefix writer writes a prefix before each new line from another io.Writer
type PrefixWriter struct {
	output              io.Writer
	outputPrefix        []byte
	outputPrefixWritten bool
}

func NewPrefixWriter(output io.Writer, prefix []byte) *PrefixWriter {
	return &PrefixWriter{
		output:       output,
		outputPrefix: prefix,
	}
}

func NewPrefixWriterString(output io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{
		output:       output,
		outputPrefix: []byte(prefix),
	}
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {

	toWrites := bytes.SplitAfter(p, []byte{'\n'})

	for _, toWrite := range toWrites {
		if len(toWrite) <= 0 {
			continue
		}
		if !w.outputPrefixWritten {
			w.outputPrefixWritten = true
			w.output.Write(w.outputPrefix)
		}

		if bytes.Contains(toWrite, []byte{'\n'}) {
			w.outputPrefixWritten = false
		}

		var written int
		written, err = w.output.Write(toWrite)
		n += written
		if err != nil {
			return
		}
	}

	return
}
