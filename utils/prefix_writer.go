package utils

import (
	"bytes"
	"io"
)

type PrefixWriter struct {
	Writer    io.Writer
	buf       *bytes.Buffer
	readLines string
	prefix    string
	persist   string
}

func NewPrefixWriter(writer io.Writer, prefix string) *PrefixWriter {
	streamer := &PrefixWriter{
		Writer:  writer,
		buf:     bytes.NewBuffer([]byte("")),
		prefix:  prefix,
		persist: "",
	}
	return streamer
}

func (l *PrefixWriter) Write(p []byte) (n int, err error) {
	if n, err = l.buf.Write(p); err != nil {
		return
	}
	err = l.OutputLines()
	return
}

func (l *PrefixWriter) Close() error {
	l.Flush()
	l.buf = bytes.NewBuffer([]byte(""))
	return nil
}

func (l *PrefixWriter) Flush() error {
	var p []byte
	if _, err := l.buf.Read(p); err != nil {
		return err
	}

	l.out(string(p))
	return nil
}

func (l *PrefixWriter) OutputLines() (err error) {
	for {
		line, err := l.buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		l.out(line)
	}
	return nil
}

func (l *PrefixWriter) FlushRecord() string {
	buffer := l.persist
	l.persist = ""
	return buffer
}

func (l *PrefixWriter) out(str string) (err error) {
	str = l.prefix + str
	_, err = l.Writer.Write([]byte(str))
	return err
}
