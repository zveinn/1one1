package main

import (
	"log"

	gonet "github.com/shirou/gopsutil/net"
)

func main() {

	var totalSent uint64
	var totalOut uint64

	netstuff, err := gonet.IOCounters(true)
	log.Println(err)
	for _, v := range netstuff {
		totalSent = totalSent + v.BytesSent
		totalOut = totalOut + v.BytesRecv
	}
	log.Println(totalSent, totalOut)
}
