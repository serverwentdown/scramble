package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/armon/go-socks5"
)

var keyInt int
var listen string
var connect string

func main() {
	flag.IntVar(&keyInt, "key", 170, "key to xor the data")
	flag.StringVar(&listen, "listen", ":8081", "listen on IP and port")
	flag.StringVar(&connect, "connect", "socks", "forward to IP and port. 'socks' sets up a SOCKS5 proxy.")
	flag.Parse()

	if keyInt < 0 || keyInt > 255 {
		flag.PrintDefaults()
		log.Fatal(fmt.Errorf("key is not one byte"))
	}
	key := byte(keyInt)

	var socks *socks5.Server
	if connect == "socks" {
		conf := &socks5.Config{}
		var err error
		socks, err = socks5.New(conf)
		if err != nil {
			log.Fatal(fmt.Errorf("unable to create socks server: %w", err))
		}
	}

	// check and parse address
	connAddr, err := net.ResolveTCPAddr("tcp", connect)
	if socks == nil && err != nil {
		flag.PrintDefaults()
		log.Fatal(fmt.Errorf("invalid connect address %s: %w", connect, err))
	}

	// listen on address
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		flag.PrintDefaults()
		log.Fatal(fmt.Errorf("unable to listen on %s: %w", listen, err))
	}

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatal(fmt.Errorf("unable to accept connection: %w", err))
		}

		log.Printf("connection from %v", c.RemoteAddr())

		scrambleConn := NewScrambleConn(c, key)
		if socks != nil {
			go socks.ServeConn(scrambleConn)
		} else {
			conn, err := net.DialTCP("tcp", nil, connAddr)
			if err != nil {
				c.Close()
				log.Print(fmt.Errorf("unable to connect to %v: %w", connAddr, err))
				continue
			}

			result := Pipe(conn, scrambleConn)
			go func() {
				pipeResult := <-result
				log.Printf("in: %d %v", pipeResult.Ingress.N, pipeResult.Ingress.Error)
				log.Printf("eg: %d %v", pipeResult.Egress.N, pipeResult.Egress.Error)
			}()
		}
	}

	log.Printf("listening on %v", ln.Addr())
}

type CloseIndividual interface {
	CloseRead() error
	CloseWrite() error
}

type PipeResult struct {
	Ingress CopyResult
	Egress  CopyResult
}

func Pipe(a, b io.ReadWriteCloser) chan PipeResult {
	// Copy from b to a
	ingressResult := Copy(a, b)
	// Copy from a to b
	egressResult := Copy(b, a)

	result := make(chan PipeResult)
	go func() {
		var in CopyResult
		var eg CopyResult
		select {
		case in = <-ingressResult:
			// b returned error
			// TODO: Consider error handling
			closeOneSide(a, b)
			eg = <-egressResult
		case eg = <-egressResult:
			// a returned error
			// TODO: Consider error handling
			closeOneSide(b, a)
			in = <-ingressResult
		}

		result <- PipeResult{
			Ingress: in,
			Egress:  eg,
		}
	}()
	return result
}

func closeOneSide(a, b io.ReadWriteCloser) (aErr error, bErr error) {
	if c, ok := a.(CloseIndividual); ok {
		aErr = c.CloseWrite()
	} else {
		aErr = a.Close()
	}
	if c, ok := b.(CloseIndividual); ok {
		bErr = c.CloseRead()
	} else {
		bErr = b.Close()
	}
	return
}

type CopyResult struct {
	N     int64
	Error error
}

func Copy(w io.Writer, r io.Reader) chan CopyResult {
	result := make(chan CopyResult)
	go func() {
		// Do a copy
		n, err := io.Copy(w, r)
		result <- CopyResult{n, err}
	}()
	return result
}
