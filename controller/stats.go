package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"log"
)

func getData(headerValue int, data []byte, valuePointer int) (index, size int, value interface{}) {
	index, size = findOrderAndSize(headerValue)
	binaryValue := data[valuePointer+1 : valuePointer+size]
	postOrNeg := data[valuePointer]
	if size == 3 {
		value := binary.LittleEndian.Uint16(binaryValue)
		if postOrNeg == 0 {
			log.Println("index/size", index, "/", size, "value: ", -value, "  ///  ", data[valuePointer:valuePointer+size])
		} else {
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[valuePointer:valuePointer+size])
		}

	} else if size == 5 {
		value := binary.LittleEndian.Uint32(binaryValue)
		if postOrNeg == 0 {
			log.Println("index/size", index, "/", size, "value: ", -value, "  ///  ", data[valuePointer:valuePointer+size])
		} else {
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[valuePointer:valuePointer+size])
		}

	} else if size == 9 {
		value := binary.LittleEndian.Uint64(binaryValue)
		if postOrNeg == 0 {
			log.Println("index/size", index, "/", size, "value: ", -value, "  ///  ", data[valuePointer:valuePointer+size])
		} else {
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[valuePointer:valuePointer+size])
		}

	} else {
		log.Println("Whooops!")
		value = 0
	}
	return
}
func parseSection(data []byte, previousEndingIndex int) (endIndex int) {
	if data[previousEndingIndex] == 0 {
		// log.Println("No data in section")
		return previousEndingIndex + 1
	}
	currentValueIndex := previousEndingIndex + int(data[previousEndingIndex]) + 1
	headerLength := int(data[previousEndingIndex])
	for i := 1; i <= headerLength; i++ {
		_, size, _ := getData(int(data[previousEndingIndex+i]), data, currentValueIndex)
		currentValueIndex = currentValueIndex + size
		endIndex = currentValueIndex
	}
	return
}
func parseNetworkingSection(data []byte, previousEndingIndex int) (endIndex int) {
	if data[previousEndingIndex] == 0 {
		// log.Println("No data in network section")
		return previousEndingIndex + 1
	}
	numberOfinterfaces := int(data[previousEndingIndex])
	currentPointer := previousEndingIndex + 1
	for i := 1; i <= numberOfinterfaces; i++ {
		currentHeaderLength := int(data[currentPointer])
		currentPointer = currentPointer + 1
		currentHeaderPointer := currentPointer
		currentPointer = currentPointer + currentHeaderLength
		for ii := 1; ii <= currentHeaderLength; ii++ {

			if ii == 1 {
				value := string(data[currentPointer : currentPointer+int(data[currentHeaderPointer])])

				log.Println("INTERFACE:", value, "  ///  ", data[currentPointer:currentPointer+int(data[currentHeaderPointer])])
				currentPointer = currentPointer + int(data[currentHeaderPointer])
				currentHeaderPointer++
			} else {
				_, size, _ := getData(int(data[currentHeaderPointer]), data, currentPointer)
				currentPointer = currentPointer + size
				currentHeaderPointer++
			}
		}
	}
	endIndex = currentPointer
	return
}
func ParseDataPoint(data []byte) {
	// log.Pri
	log.Println(data)
	log.Println("length of original:", len(data))
	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	w.Write(data)
	w.Close()
	// log.Println(b)
	log.Println("length of zlib:", b.Len())

	var bx bytes.Buffer
	w = zlib.NewWriter(&bx)
	w.Write(data)
	w.Close()
	// log.Println(bx)
	log.Println("length of flate:", bx.Len())

	log.Println("DISK:")
	diskEndIndex := parseSection(data, 0)
	// os.Exit(1)
	log.Println("MEMORY:")
	memoryEndIndex := parseSection(data, diskEndIndex)
	// memoryEndIndex = 0
	log.Println("LOAD:")
	loadEndIndex := parseSection(data, memoryEndIndex)
	log.Println("ENTROPY:")
	entropyEndIndex := parseSection(data, loadEndIndex)

	log.Println("NEWORKING:")
	networkEndIndex := parseNetworkingSection(data, entropyEndIndex)
	log.Println(networkEndIndex)
	// os.Exit(1)

}
func findOrderAndSize(data int) (index int, size int) {

	if data < 100 {

		index = data / 10
		size = data - (index * 10)
		// log.Println("data:", data, "less then 100", index, size)
	} else if data > 100 {
		index = data / 10
		size = data - (index * 10)
		// log.Println("data:", data, "more then 100", index, size)
	}

	return
}
