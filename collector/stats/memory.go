package stats

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/mem"
	"github.com/zkynetio/lynx/helpers"
)

type MemoryStatic struct {
	Total     uint64
	SwapTotal uint64
}
type MemoryDynamic struct {
	Used       uint64
	Free       uint64
	Shared     uint64
	Bufferes   uint64
	SwapFree   uint64
	Cached     uint64
	Available  uint64
	SwapCached uint64
	Percentage float64
}

func collectMemory(dp *DynamicPoint) {
	vmStat, err := mem.VirtualMemory()
	helpers.PanicX(err)

	memoryDynamicPoint := MemoryDynamic{}
	//memoryDynamicPoint.Total = vmStat.Total
	memoryDynamicPoint.Available = vmStat.Available
	memoryDynamicPoint.Used = vmStat.Used
	memoryDynamicPoint.Free = vmStat.Free
	memoryDynamicPoint.Shared = vmStat.Shared
	memoryDynamicPoint.Bufferes = vmStat.Buffers
	memoryDynamicPoint.SwapFree = vmStat.SwapFree
	memoryDynamicPoint.SwapCached = vmStat.SwapCached
	//memoryDynamicPoint.SwapTotal = vmStat.SwapTotal
	memoryDynamicPoint.Percentage = float64(vmStat.Used) / float64(vmStat.Total)

	//log.Println(memoryDynamicPoint)
	dp.MemoryDynamic = memoryDynamicPoint
}

func (m *MemoryDynamic) GetFormattedString() string {
	//helpers.DebugLog("comparing ...")
	//helpers.DebugLog(m)
	//helpers.DebugLog(History.DynamicPointMap[HighestHistoryIndex-1].Memory)

	var memorySlice []string

	memorySlice = append(memorySlice, strconv.FormatFloat(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Percentage, 'f', 6, 64))

	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Available != m.Available {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Available-m.Available)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Used != m.Used {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Used-m.Used)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Free != m.Free {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Free-m.Free)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Shared != m.Shared {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Shared-m.Shared)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Bufferes != m.Bufferes {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.Bufferes-m.Bufferes)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.SwapFree != m.SwapFree {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.SwapFree-m.SwapFree)))
	} else {
		memorySlice = append(memorySlice, "")
	}
	if History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.SwapCached != m.SwapCached {
		memorySlice = append(memorySlice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].MemoryDynamic.SwapCached-m.SwapCached)))
	} else {
		memorySlice = append(memorySlice, "")
	}

	return strings.Join(memorySlice, ",")
}
