package main

import (
	"encoding/binary"
	"log"
)

func getData(headerValue int, data []byte, valuePointer int) (index, size int, value int64) {
	index, size = findOrderAndSize(headerValue)
	binaryValue := data[valuePointer+1 : valuePointer+size]
	postOrNeg := data[valuePointer]

	if size == 1 {
		// log.Print("BB")
		return index, 0, 0
	} else if size == 3 {
		value := binary.LittleEndian.Uint16(binaryValue)
		if postOrNeg == 0 {
			return index, size, int64(-value)
		}
		return index, size, int64(value)

	} else if size == 5 {
		value := binary.LittleEndian.Uint32(binaryValue)
		if postOrNeg == 0 {
			return index, size, int64(-value)
		}
		return index, size, int64(value)

	} else if size == 9 {
		value := binary.LittleEndian.Uint64(binaryValue)
		if postOrNeg == 0 {
			return index, size, int64(-value)
		}
		return index, size, int64(value)

	} else {
		log.Println("Whooops!")
		value = 0
	}
	return index, size, int64(value)
}
func GetNetworkDataFromSection(MainIndex int, data []byte, previousEndingIndex int) (endIndex int, dpl []*ParsedDataPoint) {
	if data[previousEndingIndex] == 0 {
		return previousEndingIndex + 1, nil
	}
	numberOfinterfaces := int(data[previousEndingIndex])
	currentPointer := previousEndingIndex + 1
	for i := 1; i <= numberOfinterfaces; i++ {
		currentHeaderLength := int(data[currentPointer])
		currentPointer = currentPointer + 1
		currentHeaderPointer := currentPointer
		currentPointer = currentPointer + currentHeaderLength
		var iftag string
		for ii := 1; ii <= currentHeaderLength; ii++ {
			dp := &ParsedDataPoint{}
			if ii == 1 {
				value := string(data[currentPointer : currentPointer+int(data[currentHeaderPointer])])
				iftag = value
				currentPointer = currentPointer + int(data[currentHeaderPointer])
				currentHeaderPointer++
			} else {
				index, size, value := getData(int(data[currentHeaderPointer]), data, currentPointer)
				dp.Tag = iftag
				dp.Value = value
				dp.SubIndex = index
				dpl = append(dpl, dp)
				currentPointer = currentPointer + size
				currentHeaderPointer++
			}
		}
	}
	endIndex = currentPointer
	return
}
func GetDataFromSection(MainIndex int, data []byte, previousEndingIndex int) (endIndex int, dpl []*ParsedDataPoint) {
	if data[previousEndingIndex] == 0 {
		return previousEndingIndex + 1, nil
	}
	currentValueIndex := previousEndingIndex + int(data[previousEndingIndex]) + 1
	headerLength := int(data[previousEndingIndex])

	for i := 1; i <= headerLength; i++ {
		index, size, value := getData(int(data[previousEndingIndex+i]), data, currentValueIndex)
		dpl = append(dpl, &ParsedDataPoint{Index: MainIndex, SubIndex: index, Value: value})
		currentValueIndex = currentValueIndex + size
		endIndex = currentValueIndex
	}
	return
}
func ParseDataPoint(data []byte, tag string) (dpv *ParsedDataPointValues) {
	dpv = &ParsedDataPointValues{}
	// DiskValue := &Index{Tag: "disk", Index: 1}
	diskEndIndex, DiskValue := GetDataFromSection(1, data, 0)
	dpv.Values = append(dpv.Values, DiskValue...)
	// MemoryValue := &Index{Tag: "memory", Index: 2}
	memoryEndIndex, MemoryValue := GetDataFromSection(2, data, diskEndIndex)
	dpv.Values = append(dpv.Values, MemoryValue...)
	// LoadValue := &Index{Tag: "load", Index: 3}
	loadEndIndex, LoadValue := GetDataFromSection(3, data, memoryEndIndex)
	dpv.Values = append(dpv.Values, LoadValue...)
	// EntropyValue := &Index{Tag: "entropy", Index: 4}
	entropyEndIndex, EntropyValue := GetDataFromSection(4, data, loadEndIndex)
	dpv.Values = append(dpv.Values, EntropyValue...)
	// NetworkValue := &Index{Tag: "network", Index: 5}
	_, NetworkValue := GetNetworkDataFromSection(5, data, entropyEndIndex)
	dpv.Values = append(dpv.Values, NetworkValue...)
	return
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
