package stats

import (
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
	StaticPointMap             map[int]*StaticPoint
	DynamicPointMap            map[int]*DynamicPoint
	DynamicUpdatePoint         *DynamicPoint
	PreviousDynamicUpdatePoint *DynamicPoint
	DynamicBasePoint           *DynamicPoint
	StaticBasePoint            *StaticPoint
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
	// History.PreviousDynamicUpdatePoint = History.DynamicUpdatePoint
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
