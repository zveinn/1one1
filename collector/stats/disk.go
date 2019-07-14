package stats

import (
	"github.com/shirou/gopsutil/disk"
	"github.com/zkynetio/lynx/helpers"
)

type DiskStatic struct {
	Serial string
	Total  uint64
	Path   string
	FSType string
}
type DiskDynamic struct {
	Total          uint64
	Free           uint64
	Used           uint64
	UsedPercentage float64
	INodesTotal    uint64
	INodesUsed     uint64
	INodesFree     uint64
	ValueList      []int64
}

func collectDiskDynamic(dp *DynamicPoint) {
	// meow := disk.Partitions()
	diskStat, err := disk.Usage("/")
	helpers.PanicX(err)

	dp.DiskDynamic = &DiskDynamic{
		Total: diskStat.Total,

		Free:           diskStat.Free,
		Used:           diskStat.Used,
		UsedPercentage: diskStat.UsedPercent,
		//Path:           diskStat.Path,
		//FSType:         diskStat.Fstype,
		INodesFree:  diskStat.InodesFree,
		INodesTotal: diskStat.InodesTotal,
		INodesUsed:  diskStat.InodesUsed,
		ValueList:   []int64{},
	}
}

func (d *DiskDynamic) GetFormattedBytes(basePoint bool) []byte {

	base := History.DynamicBasePoint.DiskDynamic
	if basePoint {
		base.ValueList = append(base.ValueList, int64(base.Total))
		base.ValueList = append(base.ValueList, int64(base.Free))
		base.ValueList = append(base.ValueList, int64(base.Used))
		base.ValueList = append(base.ValueList, int64(base.INodesTotal))
		base.ValueList = append(base.ValueList, int64(base.INodesFree))
		base.ValueList = append(base.ValueList, int64(base.INodesUsed))
		base.ValueList = append(base.ValueList, int64(base.UsedPercentage))
		return helpers.WriteValueList(base.ValueList, "")
	}
	prev := History.DynamicPreviousUpdatePoint.DiskDynamic
	d.ValueList = append(d.ValueList, int64(d.Total))
	d.ValueList = append(d.ValueList, int64(d.Free))
	d.ValueList = append(d.ValueList, int64(d.Used))
	d.ValueList = append(d.ValueList, int64(d.INodesTotal))
	d.ValueList = append(d.ValueList, int64(d.INodesFree))
	d.ValueList = append(d.ValueList, int64(d.INodesUsed))
	d.ValueList = append(d.ValueList, int64(d.UsedPercentage))
	return helpers.WriteValueList2(d.ValueList, base.ValueList, prev.ValueList, "")
}
