package stats

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/mem"
	"github.com/zkynetio/lynx/helpers"
)

type Memory struct {
	Total      uint64
	Used       uint64
	Free       uint64
	Shared     uint64
	Bufferes   uint64
	SwapFree   uint64
	Cached     uint64
	Available  uint64
	SwapCached uint64
	SwapTotal  uint64
	Percentage float64
}

func collectMemory(dp *DataPoint) {
	vmStat, err := mem.VirtualMemory()
	helpers.PanicX(err)

	memoryDataPoint := Memory{}
	memoryDataPoint.Total = vmStat.Total
	memoryDataPoint.Available = vmStat.Available
	memoryDataPoint.Used = vmStat.Used
	memoryDataPoint.Free = vmStat.Free
	memoryDataPoint.Shared = vmStat.Shared
	memoryDataPoint.Bufferes = vmStat.Buffers
	memoryDataPoint.SwapFree = vmStat.SwapFree
	memoryDataPoint.SwapCached = vmStat.SwapCached
	memoryDataPoint.SwapTotal = vmStat.SwapTotal
	memoryDataPoint.Percentage = float64(vmStat.Used) / float64(vmStat.Total)

	//log.Println(memoryDataPoint)
	dp.Memory = memoryDataPoint
}

func (m *Memory) GetFormattedString() string {
	//helpers.DebugLog("comparing ...")
	//helpers.DebugLog(m)
	//helpers.DebugLog(History.DataPointMap[HighestHistoryIndex-1].Memory)

	var memorySlice []string

	memorySlice = append(memorySlice, strconv.FormatFloat(History.DataPointMap[HighestHistoryIndex-1].Memory.Percentage, 'f', 6, 64))

	if History.DataPointMap[HighestHistoryIndex-1].Memory.Total != m.Total {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Total-m.Total)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Memory.Available != m.Available {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Available-m.Available)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DataPointMap[HighestHistoryIndex-1].Memory.Used != m.Used {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Used-m.Used)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DataPointMap[HighestHistoryIndex-1].Memory.Free != m.Free {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Free-m.Free)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DataPointMap[HighestHistoryIndex-1].Memory.Shared != m.Shared {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Shared-m.Shared)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DataPointMap[HighestHistoryIndex-1].Memory.Bufferes != m.Bufferes {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.Bufferes-m.Bufferes)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Memory.SwapFree != m.SwapFree {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.SwapFree-m.SwapFree)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Memory.SwapTotal != m.SwapTotal {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.SwapTotal-m.SwapTotal)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Memory.SwapCached != m.SwapCached {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Memory.SwapCached-m.SwapCached)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	return strings.Join(memorySlice, ",")
}
