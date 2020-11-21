package trace

import (
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

type Tracer interface {
	Close() error
	Print(vs ...interface{})
	Printf(f string, s ...interface{})
	LogPacket(s string, p []byte)
}

type traceWriter struct {
	w io.WriteCloser
}

func NewTraceWriter(w io.WriteCloser) *traceWriter {
	return &traceWriter{w}
}

func (t *traceWriter) Close() (err error) {
	if t.w != nil {
		err = t.w.Close()
	}
	return
}
func (t traceWriter) Print(vs ...interface{}) {
	if t.w != nil {
		t.w.Write([]byte(fmt.Sprintf("%s: ", time.Now().Format("2006-01-02T15:04:05.0000"))))
		for _, v := range vs {
			t.w.Write([]byte(fmt.Sprintf("%v", v)))
		}
		t.w.Write([]byte{'\n'})
	}
}

func (t traceWriter) Printf(f string, s ...interface{}) {
	if t.w != nil {
		t.Print(fmt.Sprintf(f, s...))
	}
}

func (t traceWriter) LogPacket(s string, p []byte) {
	if t.w != nil {
		t.Print(s)
		t.w.Write([]byte(hex.Dump(p)))
	}
}

type nilTracer struct{}

func NilTracer() *nilTracer                         { return &nilTracer{} }
func (nilTracer) Close() error                      { return nil }
func (nilTracer) Print(vs ...interface{})           {}
func (nilTracer) Printf(f string, s ...interface{}) {}
func (nilTracer) LogPacket(s string, p []byte)      {}
