package main

import (
	"io"
	"net"
)

type ScrambleReadWriter struct {
	ReadWriter io.ReadWriter
	Key        byte
	buffer     []byte
}

func (s *ScrambleReadWriter) Read(p []byte) (n int, err error) {
	n, err = s.ReadWriter.Read(p)
	for i := 0; i < n; i++ {
		p[i] = p[i] ^ s.Key
	}
	return n, err
}

// Write takes buffer p, performs XOR in a copy of the buffer and calls write
// on the underlying ReadWriter.
func (s *ScrambleReadWriter) Write(p []byte) (n int, err error) {
	if cap(s.buffer) < cap(p) {
		s.buffer = make([]byte, 0, cap(p))
	} else {
		s.buffer = s.buffer[:0]
	}
	for i := range p {
		s.buffer = append(s.buffer, p[i]^s.Key)
	}
	return s.ReadWriter.Write(s.buffer)
}

type ScrambleConn struct {
	net.Conn
	*ScrambleReadWriter
}

func NewScrambleConn(c net.Conn, key byte) net.Conn {
	return &ScrambleConn{
		c,
		&ScrambleReadWriter{
			ReadWriter: c,
			Key:        key,
		},
	}
}

func (s *ScrambleConn) Read(b []byte) (n int, err error) {
	return s.ScrambleReadWriter.Read(b)
}
func (s *ScrambleConn) Write(b []byte) (n int, err error) {
	return s.ScrambleReadWriter.Write(b)
}
