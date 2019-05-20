package stats

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"log"
)

var History *HistoryBuffer

const HighestHistoryIndex = 2

type DynamicPoint struct {
	MemoryDynamic
	LoadDynamic
	DiskDynamic
	NetworkDynamic
	EntropyDynamic
	CPU string
	//Host      string
	General   string
	Load1MIN  float64
	Load5MIN  float64
	Load15MIN float64
}

type StaticPoint struct {
	NetworkStatic map[string]*NetworkStatic
	HostStatic
}

type HistoryBuffer struct {
	StaticPointMap     map[int]*StaticPoint
	DynamicPointMap    map[int]*DynamicPoint
	DynamicUpdatePoint *DynamicPoint
	DynamicBasePoint   *DynamicPoint
	StaticBasePoint    *StaticPoint
}

func InitStats() {
	History = &HistoryBuffer{
		DynamicPointMap: make(map[int]*DynamicPoint),
		StaticPointMap:  make(map[int]*StaticPoint),
	}
}
func CollectBasePoint() []byte {
	//var data []byte
	log.Println("base point !")
	History.DynamicBasePoint = &DynamicPoint{}
	collectDiskDynamic(History.DynamicBasePoint)
	collectLoad(History.DynamicBasePoint)
	collectMemory(History.DynamicBasePoint)
	collectEntropy(History.DynamicBasePoint)
	collectNetworkDownloadAndUpload(History.DynamicBasePoint)
	var theBytes []byte
	theBytes = append(theBytes, History.DynamicBasePoint.DiskDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.MemoryDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.LoadDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.EntropyDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.NetworkDynamic.GetFormattedBytes(true)...)
	return theBytes

}

func CollectDynamicData() []byte {
	History.DynamicUpdatePoint = &DynamicPoint{}
	collectDiskDynamic(History.DynamicUpdatePoint)
	collectLoad(History.DynamicUpdatePoint)
	collectMemory(History.DynamicUpdatePoint)
	collectEntropy(History.DynamicUpdatePoint)
	collectNetworkDownloadAndUpload(History.DynamicUpdatePoint)
	var theBytes []byte
	theBytes = append(theBytes, History.DynamicUpdatePoint.DiskDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.MemoryDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.LoadDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.EntropyDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.NetworkDynamic.GetFormattedBytes(false)...)
	return theBytes
}

func GetStaticDataPoint() string {

	if History.StaticPointMap[HighestHistoryIndex] != nil {
		History.StaticPointMap[HighestHistoryIndex-1] = History.StaticPointMap[HighestHistoryIndex]
		//helpers.DebugLog("second index", History.DynamicPointMap[HighestHistoryIndex-1])
	} else {
		History.StaticPointMap[HighestHistoryIndex-1] = &StaticPoint{}
	}

	History.StaticPointMap[HighestHistoryIndex] = &StaticPoint{}

	collectNetworkInterfaces(History.StaticPointMap[HighestHistoryIndex])
	collectStaticHostData(History.StaticPointMap[HighestHistoryIndex])
	var data string
	data = data + History.StaticPointMap[HighestHistoryIndex].HostStatic.GetFormattedString() + ";"
	data = data + getFormattedStringForInterfaces(History.StaticPointMap[HighestHistoryIndex].NetworkStatic) + ";"
	log.Println(data)
	return data
}

func ParseDataFromDynamicPoint(data []byte) {
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

	log.Println("DISK")
	diskEndIndex := 0
	currentValueIndex := int(data[0]) + 1
	startingIndex := int(data[0])
	if data[0] != 0 {
		for i := 1; i <= startingIndex; i++ {

			index, size := findOrderAndSize(int(data[0+i]))
			if size == 3 {
				value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			currentValueIndex = currentValueIndex + size
			diskEndIndex = currentValueIndex
			// log.Println("Inner loop index is:", i)
		}

	} else {
		log.Println("NO DISK!")
		diskEndIndex = 1

	}
	memoryEndIndex := 0
	if data[diskEndIndex] != 0 {
		log.Println("MEMORY")

		loopIndex := int(data[diskEndIndex])
		currentValueIndex = diskEndIndex + int(data[diskEndIndex]) + 1
		log.Println("disk end", diskEndIndex)
		log.Println("disk end bytes", data[diskEndIndex])
		log.Println("valueindex", currentValueIndex)
		log.Println("loop index:", loopIndex)
		for i := 1; i <= loopIndex; i++ {

			index, size := findOrderAndSize(int(data[diskEndIndex+i]))

			if size == 3 {
				value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			currentValueIndex = currentValueIndex + size
			memoryEndIndex = currentValueIndex
			// log.Println("Inner loop index is:", i)
		}
	} else {
		log.Println("NO MEMORY!!")
		memoryEndIndex = diskEndIndex + 1
	}

	loadEndIndex := 0
	if data[memoryEndIndex] != 0 {
		log.Println("LOAD")
		loopIndex := int(data[memoryEndIndex])
		currentValueIndex = memoryEndIndex + int(data[memoryEndIndex]) + 1
		log.Println("disk end", memoryEndIndex)
		log.Println("disk end bytes", data[memoryEndIndex])
		log.Println("valueindex", currentValueIndex)
		log.Println("loop index:", loopIndex)
		for i := 1; i <= loopIndex; i++ {

			index, size := findOrderAndSize(int(data[memoryEndIndex+i]))

			if size == 3 {
				value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			currentValueIndex = currentValueIndex + size

			// log.Println("Inner loop index is:", i)
		}
		loadEndIndex = currentValueIndex
	} else {
		log.Println("NO LOAD")
		loadEndIndex = memoryEndIndex + 1
	}

	entropyEndIndex := 0
	if data[loadEndIndex] != 0 {
		log.Println("Entropy")
		loopIndex := int(data[loadEndIndex])
		currentValueIndex = loadEndIndex + int(data[loadEndIndex]) + 1
		log.Println("disk end", loadEndIndex)
		log.Println("disk end bytes", data[loadEndIndex])
		log.Println("valueindex", currentValueIndex)
		log.Println("loop index:", loopIndex)
		for i := 1; i <= loopIndex; i++ {

			index, size := findOrderAndSize(int(data[loadEndIndex+i]))

			if size == 3 {
				value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			currentValueIndex = currentValueIndex + size
			// log.Println("Inner loop index is:", i)
		}
		entropyEndIndex = currentValueIndex
	} else {
		log.Println("NO ENTROPY")
		entropyEndIndex = loadEndIndex + 1
	}

	log.Println("network !!!")
	numberOfinterfaces := int(data[entropyEndIndex])
	log.Println("number of interfaces:", numberOfinterfaces)
	networkStartIndex := entropyEndIndex
	currentPointer := networkStartIndex + 1
	log.Println("network start index:", networkStartIndex)
	for i := 1; i <= numberOfinterfaces; i++ {
		// log.Println("Processing interface number:", i)
		currentHeaderLength := int(data[currentPointer])
		// log.Println("Header length:", currentHeaderLength)

		// current pointer becomes the first header
		currentPointer = currentPointer + 1

		currentHeaderPointer := currentPointer
		currentPointer = currentPointer + currentHeaderLength
		// log.Println("before NIF")
		for ii := 1; ii <= currentHeaderLength; ii++ {
			// log.Println("current header:", currentHeaderPointer, " data under header pointer:", data[currentHeaderPointer])
			// log.Println("current header index", ii)
			// log.Println("current pointer:", currentPointer)
			// log.Println(data[currentPointer])
			var size int
			var index int
			if ii == 1 {
				value := string(data[currentPointer : currentPointer+int(data[currentHeaderPointer])])

				log.Println("value: ", value, "  ///  ", data[currentPointer:currentPointer+int(data[currentHeaderPointer])])
				currentPointer = currentPointer + int(data[currentHeaderPointer])
				currentHeaderPointer++
				continue
			} else {
				index, size = findOrderAndSize(int(data[currentHeaderPointer]))
			}
			data := data[currentPointer+1 : currentPointer+size]
			if size == 3 {
				value := binary.LittleEndian.Uint16(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			// log.Println("current pointer at end:", currentPointer)
			// log.Println("Adding to current pointer:", size)
			currentPointer = currentPointer + size
			currentHeaderPointer++

		}

		// log.Println("Inner loop index is:", i)
	}

}
func ParseDataFromBasePoint(data []byte) {
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

	log.Println("DISK")
	diskEndIndex := 0
	currentValueIndex := int(data[0]) + 1
	startingIndex := int(data[0])
	for i := 1; i <= startingIndex; i++ {

		index, size := findOrderAndSize(int(data[0+i]))
		if size == 3 {
			value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 5 {
			value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 9 {
			value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		}
		// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
		currentValueIndex = currentValueIndex + size
		diskEndIndex = currentValueIndex
		// log.Println("Inner loop index is:", i)
	}

	log.Println("MEMORY")
	memoryEndIndex := 0
	loopIndex := int(data[diskEndIndex])
	currentValueIndex = diskEndIndex + int(data[diskEndIndex]) + 1
	log.Println("disk end", diskEndIndex)
	log.Println("disk end bytes", data[diskEndIndex])
	log.Println("valueindex", currentValueIndex)
	log.Println("loop index:", loopIndex)
	for i := 1; i <= loopIndex; i++ {

		index, size := findOrderAndSize(int(data[diskEndIndex+i]))

		if size == 3 {
			value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 5 {
			value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 9 {
			value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		}
		// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
		currentValueIndex = currentValueIndex + size
		memoryEndIndex = currentValueIndex
		// log.Println("Inner loop index is:", i)
	}

	log.Println("LOAD")
	loadEndIndex := 0
	loopIndex = int(data[memoryEndIndex])
	currentValueIndex = memoryEndIndex + int(data[memoryEndIndex]) + 1
	log.Println("disk end", memoryEndIndex)
	log.Println("disk end bytes", data[memoryEndIndex])
	log.Println("valueindex", currentValueIndex)
	log.Println("loop index:", loopIndex)
	for i := 1; i <= loopIndex; i++ {

		index, size := findOrderAndSize(int(data[memoryEndIndex+i]))

		if size == 3 {
			value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 5 {
			value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 9 {
			value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		}
		// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
		currentValueIndex = currentValueIndex + size

		// log.Println("Inner loop index is:", i)
	}
	loadEndIndex = currentValueIndex

	log.Println("Entropy")
	entropyEndIndex := 0
	loopIndex = int(data[loadEndIndex])
	currentValueIndex = loadEndIndex + int(data[loadEndIndex]) + 1
	log.Println("disk end", loadEndIndex)
	log.Println("disk end bytes", data[loadEndIndex])
	log.Println("valueindex", currentValueIndex)
	log.Println("loop index:", loopIndex)
	for i := 1; i <= loopIndex; i++ {

		index, size := findOrderAndSize(int(data[loadEndIndex+i]))

		if size == 3 {
			value := binary.LittleEndian.Uint16(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 5 {
			value := binary.LittleEndian.Uint32(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		} else if size == 9 {
			value := binary.LittleEndian.Uint64(data[currentValueIndex+1 : currentValueIndex+size])
			log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data[currentValueIndex+1:currentValueIndex+size])
		}
		// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
		currentValueIndex = currentValueIndex + size
		// log.Println("Inner loop index is:", i)
	}

	entropyEndIndex = currentValueIndex
	log.Println(entropyEndIndex)
	log.Println(data[entropyEndIndex:])
	log.Println("network !!!")
	numberOfinterfaces := int(data[entropyEndIndex])
	log.Println("number of interfaces:", numberOfinterfaces)
	networkStartIndex := entropyEndIndex
	currentPointer := networkStartIndex + 1
	log.Println("network start index:", networkStartIndex)
	for i := 1; i <= numberOfinterfaces; i++ {
		// log.Println("Processing interface number:", i)
		currentHeaderLength := int(data[currentPointer])
		// log.Println("Header length:", currentHeaderLength)

		// current pointer becomes the first header
		currentPointer = currentPointer + 1

		currentHeaderPointer := currentPointer
		currentPointer = currentPointer + currentHeaderLength
		// log.Println("before NIF")
		for ii := 1; ii <= currentHeaderLength; ii++ {
			// log.Println("current header:", currentHeaderPointer, " data under header pointer:", data[currentHeaderPointer])
			// log.Println("current header index", ii)
			// log.Println("current pointer:", currentPointer)
			// log.Println(data[currentPointer])
			var size int
			var index int
			if ii == 1 {
				value := string(data[currentPointer : currentPointer+int(data[currentHeaderPointer])])

				log.Println("value: ", value, "  ///  ", data[currentPointer:currentPointer+int(data[currentHeaderPointer])])
				currentPointer = currentPointer + int(data[currentHeaderPointer])
				currentHeaderPointer++
				continue
			} else {
				index, size = findOrderAndSize(int(data[currentHeaderPointer]))
			}
			data := data[currentPointer+1 : currentPointer+size]
			if size == 3 {
				value := binary.LittleEndian.Uint16(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			} else if size == 5 {
				value := binary.LittleEndian.Uint32(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			} else if size == 9 {
				value := binary.LittleEndian.Uint64(data)
				log.Println("index/size", index, "/", size, "value: ", value, "  ///  ", data)
			}
			// log.Println("Value: ", data[currentValueIndex:currentValueIndex+size])
			// log.Println("current pointer at end:", currentPointer)
			// log.Println("Adding to current pointer:", size)
			currentPointer = currentPointer + size
			currentHeaderPointer++

		}

		// log.Println("Inner loop index is:", i)
	}

}
func findOrderAndSize(data int) (index int, size int) {
	// if data >= 10 && data < 20 {
	// 	index = 1
	// 	size = data - 10
	// } else if data >= 20 && data < 30 {
	// 	index = 2
	// 	size = data - 20
	// } else if data >= 20 && data < 30 {
	// 	index = 2
	// 	size = data - 30
	// } else if data >= 30 && data < 40 {
	// 	index = 2
	// 	size = data - 50
	// } else if data >= 50 && data < 30 {
	// 	index = 2
	// 	size = data - 50
	// } else if data >= 50 && data < 30 {
	// 	index = 2
	// 	size = data - 20
	// } else if data >= 20 && data < 30 {
	// 	index = 2
	// 	size = data - 20
	// }
	// log.Println("index and size:", data)
	if data < 100 {
		index = data / 10
		size = data - (index * 10)
	} else if data > 100 {
		index = data / 10
		size = data - (index * 10)
	}

	return
}
