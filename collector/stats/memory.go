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
	Total      uint64
	Used       uint64
	Free       uint64
	Shared     uint64
	Buffers    uint64
	SwapFree   uint64
	Cached     uint64
	Available  uint64
	SwapCached uint64
	Percentage float64
	ValueList  []int64
}

func GetMemoryBytes() byte {
	vmStat, err := mem.VirtualMemory()
	helpers.PanicX(err)
	return byte(vmStat.UsedPercent)
}
func collectMemory(dp *DynamicPoint) {
	vmStat, err := mem.VirtualMemory()
	helpers.PanicX(err)

	memoryDynamicPoint := MemoryDynamic{}
	//memoryDynamicPoint.Total = vmStat.Total
	memoryDynamicPoint.Total = vmStat.Total
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
	dp.MemoryDynamic = &memoryDynamicPoint
}

func (d *MemoryDynamic) GetFormattedBytes(basePoint bool) []byte {
	base := History.DynamicBasePoint.MemoryDynamic
	if basePoint {
		base.ValueList = append(base.ValueList, int64(base.Total))
		base.ValueList = append(base.ValueList, int64(base.Free))
		base.ValueList = append(base.ValueList, int64(base.Used))
		base.ValueList = append(base.ValueList, int64(base.Available))
		base.ValueList = append(base.ValueList, int64(base.Shared))
		base.ValueList = append(base.ValueList, int64(base.Buffers))
		base.ValueList = append(base.ValueList, int64(base.SwapFree))
		base.ValueList = append(base.ValueList, int64(base.SwapCached))
		base.ValueList = append(base.ValueList, int64(base.Percentage))
		return helpers.WriteValueList(base.ValueList, "")
	}

	prev := History.DynamicPreviousUpdatePoint.MemoryDynamic
	d.ValueList = append(d.ValueList, int64(d.Total))
	d.ValueList = append(d.ValueList, int64(d.Free))
	d.ValueList = append(d.ValueList, int64(d.Used))
	d.ValueList = append(d.ValueList, int64(d.Available))
	d.ValueList = append(d.ValueList, int64(d.Shared))
	d.ValueList = append(d.ValueList, int64(d.Buffers))
	d.ValueList = append(d.ValueList, int64(d.SwapFree))
	d.ValueList = append(d.ValueList, int64(d.SwapCached))
	d.ValueList = append(d.ValueList, int64(d.Percentage))
	return helpers.WriteValueList2(d.ValueList, base.ValueList, prev.ValueList, "")

}
