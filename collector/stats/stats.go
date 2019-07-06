package stats

import (
	"log"
)

var History *HistoryBuffer

const HighestHistoryIndex = 2

type DynamicPoint struct {
	MemoryDynamic  *MemoryDynamic
	LoadDynamic    *LoadDynamic
	DiskDynamic    *DiskDynamic
	NetworkDynamic *NetworkDynamic
	EntropyDynamic *EntropyDynamic
	CPU            string
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
	DynamicPreviousUpdatePoint *DynamicPoint
	DynamicUpdatePoint         *DynamicPoint
	DynamicBasePoint           *DynamicPoint

	StaticPreviousUpdatePoint *StaticPoint
	StaticBasePoint           *StaticPoint
	StaticUpdatePoint         *StaticPoint
}

func InitStats() {
	History = &HistoryBuffer{}
}

func CollectBasePoint() []byte {
	log.Println("base point !")
	History.DynamicBasePoint = &DynamicPoint{}
	History.DynamicPreviousUpdatePoint = &DynamicPoint{}
	collectDiskDynamic(History.DynamicBasePoint)
	collectLoad(History.DynamicBasePoint)
	collectMemory(History.DynamicBasePoint)
	collectEntropy(History.DynamicBasePoint)
	collectNetworkDownloadAndUpload(History.DynamicBasePoint)
	var theBytes []byte
	// THIS ORDER MATTERS.. DO NOT CHANGE
	theBytes = append(theBytes, History.DynamicBasePoint.DiskDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.MemoryDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.LoadDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.EntropyDynamic.GetFormattedBytes(true)...)
	theBytes = append(theBytes, History.DynamicBasePoint.NetworkDynamic.GetFormattedBytes(true)...)
	History.DynamicPreviousUpdatePoint = History.DynamicBasePoint
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
	// THIS ORDER MATTERS.. DO NOT CHANGE!
	theBytes = append(theBytes, History.DynamicUpdatePoint.DiskDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.MemoryDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.LoadDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.EntropyDynamic.GetFormattedBytes(false)...)
	theBytes = append(theBytes, History.DynamicUpdatePoint.NetworkDynamic.GetFormattedBytes(false)...)
	History.DynamicPreviousUpdatePoint = History.DynamicUpdatePoint
	return theBytes
}
func GetStaticBasePoint() string {
	History.StaticBasePoint = &StaticPoint{}
	collectNetworkStats(History.StaticBasePoint)
	collectStaticHostData(History.StaticBasePoint)
	var data string
	data = data + History.StaticBasePoint.HostStatic.GetFormattedString() + ";"
	data = data + getFormattedStringForInterfaces(History.StaticBasePoint.NetworkStatic) + ";"
	History.StaticPreviousUpdatePoint = History.StaticUpdatePoint
	log.Println(data)
	return data
}

func GetStaticDataPoint() string {
	History.StaticUpdatePoint = &StaticPoint{}
	collectNetworkStats(History.StaticUpdatePoint)
	collectStaticHostData(History.StaticUpdatePoint)
	var data string
	data = data + History.StaticUpdatePoint.HostStatic.GetFormattedString() + ";"
	data = data + getFormattedStringForInterfaces(History.StaticUpdatePoint.NetworkStatic) + ";"
	log.Println(data)
	History.StaticPreviousUpdatePoint = History.StaticUpdatePoint
	return data
}

func CheckStaticDataForChanges() string {
	// TODO: find a way to do a deep compare on structs
	// --- maybe you can do this with reflect
	// TODO: send only the changed parts..
	var hasChanged bool
	if History.StaticUpdatePoint.HostID != History.StaticBasePoint.HostID {
		log.Println("static data change !")
		hasChanged = true
	}

	log.Println(hasChanged)
	return ""
}
