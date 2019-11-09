package controller

import (
	"encoding/binary"
	"log"
)

type DP struct {
	Index int
	Value int
}
type DPCollection struct {
	Tag         string `json:"tag"`
	DPS         []DP   `json:"dps"`
	Timestamp   uint64 `json:"-"`
	ControlByte int    `json:"-"`
}

func ParseMinimumDataPoint(data []byte, namespaces map[int]string) DPCollection {
	log.Println("FULL DATA", data)
	cpu := int8(data[0])
	disk := int8(data[1])
	memory := int8(data[2])
	// networkIN := binary.LittleEndian.Uint64(data[3:11])
	// networkOUT := binary.LittleEndian.Uint64(data[11:19])
	log.Println("CPU:", cpu)
	log.Println("DISK:", disk)
	log.Println("MEMORY:", memory)
	// log.Println("NETWORK IN:", networkIN)
	// log.Println("NETWORK OUT:", networkOUT)

	currentDataIndex := 0
	DPC := DPCollection{}
	if _, ok := namespaces[1]; ok {
		DPC.DPS = append(DPC.DPS, DP{Value: int(data[currentDataIndex]), Index: 1})
		currentDataIndex++
	}
	if _, ok := namespaces[2]; ok {
		DPC.DPS = append(DPC.DPS, DP{Value: int(data[currentDataIndex]), Index: 2})
		currentDataIndex++
	}
	if _, ok := namespaces[3]; ok {
		DPC.DPS = append(DPC.DPS, DP{Value: int(data[currentDataIndex]), Index: 3})
		currentDataIndex++
	}

	if _, ok := namespaces[4]; ok {
		lengthOfNetworkIn := int(data[currentDataIndex])
		// log.Println("length of network in", lengthOfNetworkIn)
		dp := DP{Index: 4}
		if lengthOfNetworkIn == 2 {
			dp.Value = int(binary.LittleEndian.Uint16(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkIn]))
		} else if lengthOfNetworkIn == 4 {
			dp.Value = int(binary.LittleEndian.Uint32(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkIn]))
		} else if lengthOfNetworkIn == 8 {
			dp.Value = int(binary.LittleEndian.Uint64(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkIn]))
		}

		DPC.DPS = append(DPC.DPS, dp)
		currentDataIndex = currentDataIndex + 1 + lengthOfNetworkIn
		// log.Println("Data index post NI:", currentDataIndex)
	}

	if _, ok := namespaces[4]; ok {
		lengthOfNetworkOut := int(data[currentDataIndex])
		dp := DP{Index: 5}
		if lengthOfNetworkOut == 2 {
			dp.Value = int(binary.LittleEndian.Uint16(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkOut]))
		} else if lengthOfNetworkOut == 4 {
			dp.Value = int(binary.LittleEndian.Uint32(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkOut]))
		} else if lengthOfNetworkOut == 8 {
			dp.Value = int(binary.LittleEndian.Uint64(data[currentDataIndex+1 : currentDataIndex+1+lengthOfNetworkOut]))
		}

		DPC.DPS = append(DPC.DPS, dp)
		currentDataIndex = currentDataIndex + lengthOfNetworkOut
		// log.Println("Data index post NO:", currentDataIndex)
	}

	return DPC
}
