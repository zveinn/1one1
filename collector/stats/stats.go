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
func CollectBasePoint() string {
	//var data []byte
	History.DynamicBasePoint = &DynamicPoint{}
	collectNetworkDownloadAndUpload(History.DynamicBasePoint)
	collectLoad(History.DynamicBasePoint)
	collectMemory(History.DynamicBasePoint)
	collectDiskDynamic(History.DynamicBasePoint)
	collectEntropy(History.DynamicBasePoint)
	return ""
}

func CollectDynamicData() string {
	History.DynamicPointMap[HighestHistoryIndex-1] = History.DynamicPointMap[HighestHistoryIndex]
	History.DynamicPointMap[HighestHistoryIndex] = &DynamicPoint{}
	// log.Println(time.Now().Sub(*lastBasePoint))
	History.DynamicUpdatePoint = &DynamicPoint{}
	collectNetworkDownloadAndUpload(History.DynamicPointMap[HighestHistoryIndex])
	// new
	collectDiskDynamic(History.DynamicUpdatePoint)
	collectLoad(History.DynamicUpdatePoint)
	collectMemory(History.DynamicUpdatePoint)
	collectEntropy(History.DynamicUpdatePoint)
	var theBytes []byte
	theBytes = append(theBytes, History.DynamicUpdatePoint.DiskDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.MemoryDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.LoadDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.EntropyDynamic.GetFormattedBytes(false)...)
	log.Println(theBytes)
	// old

	var data string
	data = data + History.DynamicPointMap[HighestHistoryIndex].NetworkDynamic.GetFormattedString()

	return data
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
