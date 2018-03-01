package main

import (
	"flag"
	"io"
	"log"
	"net"
)

var key int
var listen string
var connect string

func main() {
	flag.IntVar(&key, "key", 170, "key to xor the data")
	flag.StringVar(&listen, "listen", ":8081", "listen on ip and port")
	flag.StringVar(&connect, "connect", ":8080", "forward to ip and port")
	flag.Parse()

	if key < 0 || key > 255 {
		flag.PrintDefaults()
		log.Fatal("key is not one byte")
	}

	// check and parse address
	conn, err := net.ResolveTCPAddr("tcp", connect)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal(err)
	}

	// listen on address
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal(err)
	}

	log.Printf("listening on %v", ln.Addr())
	log.Printf("will connect to %v", conn)

	for i := 0; ; i++ {
		// accept new connection
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("connection %v from %v", i, c.RemoteAddr())

		cn, err := net.DialTCP("tcp", nil, conn)
		if err != nil {
			c.Close()
			log.Print(err)
			continue
		}

		go pipe(c, cn, byte(key), i)
		go pipe(cn, c, byte(key), i)
	}
}

func pipe(w io.WriteCloser, r io.ReadCloser, key byte, count int) {
	n, err := copyBufferXor(w, r, key)

	r.Close()
	w.Close()

	log.Printf("connection %v closed, %v bytes", count, n)

	opError, ok := err.(*net.OpError)
	if err != nil && (!ok || opError.Op != "readfrom") {
		log.Printf("warning! %v", err)
	}
}

func copyBufferXor(dst io.Writer, src io.Reader, key byte) (written int64, err error) {
	buf := make([]byte, 32*1024)
	for {
		nr, er := src.Read(buf)
		for i := 0; i < nr; i++ {
			buf[i] = buf[i] ^ key
		}
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
