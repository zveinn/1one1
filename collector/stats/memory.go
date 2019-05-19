package stats

import (
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
	Buffers    uint64
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
	memoryDynamicPoint.Buffers = vmStat.Buffers
	memoryDynamicPoint.SwapFree = vmStat.SwapFree
	memoryDynamicPoint.SwapCached = vmStat.SwapCached
	//memoryDynamicPoint.SwapTotal = vmStat.SwapTotal
	memoryDynamicPoint.Percentage = float64(vmStat.Used) / float64(vmStat.Total)

	//log.Println(memoryDynamicPoint)
	dp.MemoryDynamic = memoryDynamicPoint
}

func (d *MemoryDynamic) GetFormattedBytes(basePoint bool) []byte {

	var valueList []int64
	base := History.DynamicBasePoint.MemoryDynamic
	if basePoint {
		valueList = append(valueList, int64(base.Free))
		valueList = append(valueList, int64(base.Used))
		valueList = append(valueList, int64(base.Available))
		valueList = append(valueList, int64(base.Shared))
		valueList = append(valueList, int64(base.Buffers))
		valueList = append(valueList, int64(base.SwapFree))
		valueList = append(valueList, int64(base.SwapCached))
		valueList = append(valueList, int64(base.Percentage))
	} else {
		valueList = append(valueList, int64(d.Free)-int64(base.Free))
		valueList = append(valueList, int64(d.Used)-int64(base.Used))
		valueList = append(valueList, int64(d.Available)-int64(base.Available))
		valueList = append(valueList, int64(d.Shared)-int64(base.Shared))
		valueList = append(valueList, int64(d.Buffers)-int64(base.Buffers))
		valueList = append(valueList, int64(d.SwapFree)-int64(base.SwapFree))
		valueList = append(valueList, int64(d.SwapCached)-int64(base.SwapCached))
		valueList = append(valueList, int64(d.Percentage)-int64(base.Percentage))
	}

	return helpers.WriteValueList(valueList, "")
}
