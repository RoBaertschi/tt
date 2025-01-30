package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"robaertschi.xyz/robaertschi/tt/term"
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

//go:generate stringer -type=Level -linecomment
type Level int

const (
	Debug Level = iota // DEBUG
	Info               // INFO
	Warn               // WARN
	Error              // ERROR
	Fatal              // FATAL
)

type LoggerFormatFunc func(prefix string, level Level, msg string) string

type Logger struct {
	outMu  sync.Mutex
	out    io.Writer
	prefix string
	format LoggerFormatFunc
	filter Level
}

func DefaultLoggerFormatFunc(prefix string, level Level, msg string) string {

	colorString := ""
	switch level {
	case Debug:
		colorString = term.CSI + "90m"
	case Info:
	case Warn:
		colorString = term.Color(term.YellowFg)
	case Error:
		colorString = term.Color(term.RedFg)
	case Fatal:
		colorString = term.CSI + term.WhiteFg + term.RedBg + "m"
	}

	return fmt.Sprintf("%s%s[%s] %s%s", colorString, prefix, level, msg, term.Reset)
}

// filter filters anything below that level out, it does not stop a os.Exit() from a fatal
func NewLogger(output io.Writer, prefix string, filter Level) *Logger {
	l := new(Logger)
	l.SetPrefix(prefix)
	l.SetOutput(output)
	l.format = DefaultLoggerFormatFunc
	l.filter = filter
	return l
}

func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

func (l *Logger) SetOutput(output io.Writer) {
	// NOTE(Robin): Do some research/testing if we need to look the mutex for this
	l.out = output
}

func (l *Logger) Msg(level Level, msg string) {
	if level >= l.filter {
		l.outMu.Lock()
		result := l.format(l.prefix, level, strings.TrimRight(msg, " \n\t"))
		io.WriteString(l.out, result)
		io.WriteString(l.out, "\n")
		l.outMu.Unlock()
	}

	if level == Fatal {
		term.Exit(1)
	}
}

func (l *Logger) Msgf(level Level, msg string, args ...any) {
	l.Msg(level, fmt.Sprintf(msg, args...))
}

func (l *Logger) Debug(msg string) {
	l.Msg(Debug, msg)
}

func (l *Logger) Debugf(msg string, args ...any) {
	l.Msgf(Debug, msg, args...)
}

func (l *Logger) Info(msg string) {
	l.Msg(Info, msg)
}

func (l *Logger) Infof(msg string, args ...any) {
	l.Msgf(Info, msg, args...)
}

func (l *Logger) Warn(msg string) {
	l.Msg(Warn, msg)
}

func (l *Logger) Warnf(msg string, args ...any) {
	l.Msgf(Warn, msg, args...)
}

func (l *Logger) Error(msg string) {
	l.Msg(Error, msg)
}

func (l *Logger) Errorf(msg string, args ...any) {
	l.Msgf(Error, msg, args...)
}

func (l *Logger) Fatal(msg string) {
	l.Msg(Fatal, msg)
}

func (l *Logger) Fatalf(msg string, args ...any) {
	l.Msgf(Fatal, msg, args...)
}
