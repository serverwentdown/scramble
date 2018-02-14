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
	flag.StringVar(&listen, "listen", ":8000", "listen on ip and port")
	flag.StringVar(&connect, "connect", "", "forward to ip and port")
	flag.Parse()

	if key < 0 || key > 255 {
        flag.PrintDefaults()
		log.Fatal("key is not one byte")
	}

    if len(connect) < 3 {
        flag.PrintDefaults()
        log.Fatal("no connection specified")
    }

	ln, err := net.Listen("tcp", listen)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal(err)
	}
	log.Print("listening on " + listen)
	log.Print("will connect to " + connect)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Print(err)
            continue
		}	
        log.Print("Connection from "+c.RemoteAddr().String())

		cn, err := net.Dial("tcp", connect)
		if err != nil {
			log.Print(err)
            continue
		}

		go pipe(c, cn, byte(key))
		go pipe(cn, c, byte(key))
	}
}

func pipe(w io.WriteCloser, r io.ReadCloser, key byte) {
	buff := make([]byte, 65535)
	for {
		n, rerr := r.Read(buff)
		for i := 0; i < n; i++ {
			buff[i] = buff[i] ^ key
		}
		wn, werr := w.Write(buff[:n])
        if n != wn {
            log.Print("mismatch")
        }

		if werr != nil {
			r.Close()
			log.Print(werr)
            return
		}
		if rerr != nil {
			w.Close()
			log.Print(rerr)
            return
		}
	}
}
