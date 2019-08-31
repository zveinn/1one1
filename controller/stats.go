package controller

import (
	"encoding/binary"
	"log"
)

type ParsedDataPointValues struct {
	Values []*ParsedDataPoint
}
type ParsedCollection struct {
	Tag       string
	DPS       []*ParsedDataPoint
	BasePoint bool
}
type ParsedDataPoint struct {
	SubIndex int
	Index    int
	Tag      string
	Value    int64
	Group    string
}

func getData(headerValue int, data []byte, valuePointer int) (index, size int, value int64) {
	index, size = findOrderAndSize(headerValue)
	if size == 1 {
		// log.Print("BB")
		return index, 0, 0
	}
	binaryValue := data[valuePointer+1 : valuePointer+size]
	postOrNeg := data[valuePointer]
	if size == 3 {
		value := binary.LittleEndian.Uint16(binaryValue)
		if postOrNeg == 0 {
			// log.Println("WAS NEGATIVE", -int64(value))
			return index, size, -int64(value)
		}
		return index, size, int64(value)

	} else if size == 5 {
		value := binary.LittleEndian.Uint32(binaryValue)
		if postOrNeg == 0 {
			// log.Println("WAS NEGATIVE", -int64(value))
			return index, size, -int64(value)
		}
		return index, size, int64(value)

	} else if size == 9 {
		value := binary.LittleEndian.Uint64(binaryValue)
		if postOrNeg == 0 {
			// log.Println("WAS NEGATIVE:", -int64(value))
			return index, size, -int64(value)
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
				dp.Index = MainIndex
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

func ParseMinimumDataPoint(data []byte) DPCollection {
	// log.Println("FULL DATA", data)
	// cpu := int8(data[0])
	// disk := int8(data[1])
	// memory := int8(data[2])
	// networkIN := binary.LittleEndian.Uint64(data[3:11])
	// networkOUT := binary.LittleEndian.Uint64(data[11:19])
	// log.Println("CPU:", cpu)
	// log.Println("DISK:", disk)
	// log.Println("MEMORY:", memory)
	// log.Println("NETWORK IN:", networkIN)
	// log.Println("NETWORK OUT:", networkOUT)
	DPC := DPCollection{}
	DPC.DPS = append(DPC.DPS, DP{Value: int(data[0]), Index: 1})
	DPC.DPS = append(DPC.DPS, DP{Value: int(data[1]), Index: 2})
	DPC.DPS = append(DPC.DPS, DP{Value: int(data[2]), Index: 3})
	DPC.DPS = append(DPC.DPS, DP{Value: int(binary.LittleEndian.Uint64(data[3:11])), Index: 4})
	DPC.DPS = append(DPC.DPS, DP{Value: int(binary.LittleEndian.Uint64(data[11:19])), Index: 5})
	return DPC
}
func ParseDataPoint(data []byte, tag string) (dpv *ParsedDataPointValues) {
	dpv = &ParsedDataPointValues{}
	// log.Println(data)
	// log.Print(1)
	diskEndIndex, DiskValue := GetDataFromSection(1, data, 0)
	dpv.Values = append(dpv.Values, DiskValue...)
	// log.Print(2, diskEndIndex)
	memoryEndIndex, MemoryValue := GetDataFromSection(2, data, diskEndIndex)
	dpv.Values = append(dpv.Values, MemoryValue...)
	// log.Print(3, memoryEndIndex)
	loadEndIndex, LoadValue := GetDataFromSection(3, data, memoryEndIndex)
	dpv.Values = append(dpv.Values, LoadValue...)
	// log.Print(4, loadEndIndex)
	entropyEndIndex, EntropyValue := GetDataFromSection(4, data, loadEndIndex)
	dpv.Values = append(dpv.Values, EntropyValue...)
	// log.Print(5, entropyEndIndex)
	networkEndIndex, NetworkValue := GetNetworkDataFromSection(5, data, entropyEndIndex)
	dpv.Values = append(dpv.Values, NetworkValue...)
	// log.Print(6, networkEndIndex)
	_, CPUVaues := GetDataFromSection(6, data, networkEndIndex)
	dpv.Values = append(dpv.Values, CPUVaues...)
	// log.Print(7, cpuEndIndex)
	// for _, v := range dpv.Values {
	// 	log.Println(v.Index, v.Value)
	// }
	// log.Println(dpv.Values)
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
