package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"unsafe"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:3333")
	for {
		if err != nil {
			panic(err)
		}
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		doConnectionThing(conn)
	}

}
func doConnectionThing(conn net.Conn) {

	xx2 := float64(128)
	xx := int8(xx2)
	mepw := unsafe.Sizeof(xx)
	log.Println(mepw)
	log.Println(xx)
	scanner := bufio.NewScanner(conn)
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		index := bytes.Index(data, []byte{0, 0, 0, 0})
		log.Println(index)
		if index != -1 {
			return index + 4, data[:index], nil
		}
		if atEOF {
			return 0, nil, io.EOF
		}
		return 0, nil, nil
	}
	scanner.Split(split)
	for scanner.Scan() {
		log.Println(scanner.Bytes())
	}

	if scanner.Err() != nil {
		log.Println(scanner.Err())
	}
}
