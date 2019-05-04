package main

import "net"

func main() {

	// connect to this socket
	conn, _ := net.Dial("tcp", "127.0.0.1:3333")

	conn.Write([]byte{55, 1, 1, 0, 0, 1, 1, 1, 1, 1, 1, 255, 0, 0, 0, 0, 22, 43, 34, 34, 34, 34, 34, 34, 224, 0, 0, 0, 0})
}
