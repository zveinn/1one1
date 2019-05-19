package stats

import (
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
func ParseData(data []byte) {
	// log.Println(data)

	log.Println("Data length LENGTH:", len(data))

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

	if data < 100 {
		index = data / 10
		size = data - (index * 10)
	} else if data > 100 {
		index = data / 100
		size = data - (index * 10)
	}

	return
}
