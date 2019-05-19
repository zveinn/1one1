package main

import (
	"log"
	"strconv"
)

func main() {

	lengthstring := strconv.Itoa(4 + 1)
	index := strconv.Itoa(10)
	// final := index + lengthstring

	final, _ := strconv.Atoi(index + lengthstring)
	log.Println(byte(final))
	log.Println(final)
	buf := make([]byte, 10)
	buf = append(buf, []byte{byte(final)}...)
	// connect to this socket
	log.Println(buf)
}
