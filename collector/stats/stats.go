package stats

import (
	"log"
)

var History *HistoryBuffer

const HighestHistoryIndex = 5

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
	StaticPointMap  map[int]*StaticPoint
	DynamicPointMap map[int]*DynamicPoint
}

func InitStats() {
	History = &HistoryBuffer{
		DynamicPointMap: make(map[int]*DynamicPoint),
		StaticPointMap:  make(map[int]*StaticPoint),
	}
}
func CollectDynamicData2() []byte {
	//var data []byte

	return nil
}
func CollectDynamicData() string {
	// move DynamicPoints
	// TODO: do better
	if History.DynamicPointMap[HighestHistoryIndex] != nil {
		History.DynamicPointMap[HighestHistoryIndex-1] = History.DynamicPointMap[HighestHistoryIndex]
		//helpers.DebugLog("second index", History.DynamicPointMap[HighestHistoryIndex-1])
	} else {
		History.DynamicPointMap[HighestHistoryIndex-1] = &DynamicPoint{}
	}

	History.DynamicPointMap[HighestHistoryIndex] = &DynamicPoint{}
	//helpers.DebugLog("current index", History.DynamicPointMap[HighestHistoryIndex])
	//helpers.DebugLog("second index", History.DynamicPointMap[HighestHistoryIndex-1])

	// start
	collectNetworkDownloadAndUpload(History.DynamicPointMap[HighestHistoryIndex])
	collectLoad(History.DynamicPointMap[HighestHistoryIndex])
	collectMemory(History.DynamicPointMap[HighestHistoryIndex])
	collectDiskDynamic(History.DynamicPointMap[HighestHistoryIndex])
	collectEntropy(History.DynamicPointMap[HighestHistoryIndex])
	_ = History.DynamicPointMap[HighestHistoryIndex].DiskDynamic.GetFormattedBytes()
	var data string
	data = History.DynamicPointMap[HighestHistoryIndex].MemoryDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].LoadDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].DiskDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].EntropyDynamic.GetFormattedString() + ";"
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
