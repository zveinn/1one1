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
	Free           uint64
	Used           uint64
	UsedPercentage float64
	INodesTotal    uint64
	INodesUsed     uint64
	INodesFree     uint64
}

func collectDiskDynamic(dp *DynamicPoint) {
	diskStat, err := disk.Usage("/")
	helpers.PanicX(err)

	dp.DiskDynamic = DiskDynamic{
		//Total:          diskStat.Total,
		Free:           diskStat.Free,
		Used:           diskStat.Used,
		UsedPercentage: diskStat.UsedPercent,
		//Path:           diskStat.Path,
		//FSType:         diskStat.Fstype,
		INodesFree:  diskStat.InodesFree,
		INodesTotal: diskStat.InodesTotal,
		INodesUsed:  diskStat.InodesUsed,
	}
}

func (d *DiskDynamic) GetFormattedBytes(basePoint bool) []byte {

	var valueList []int64
	base := History.DynamicBasePoint.DiskDynamic
	if basePoint {
		valueList = append(valueList, int64(base.Free))
		valueList = append(valueList, int64(base.Used))
		valueList = append(valueList, int64(base.INodesTotal))
		valueList = append(valueList, int64(base.INodesFree))
		valueList = append(valueList, int64(base.INodesUsed))
		valueList = append(valueList, int64(base.UsedPercentage))
	} else {
		valueList = append(valueList, int64(d.Free)-int64(base.Free))
		valueList = append(valueList, int64(d.Used)-int64(base.Used))
		valueList = append(valueList, int64(d.INodesTotal)-int64(base.INodesTotal))
		valueList = append(valueList, int64(d.INodesFree)-int64(base.INodesFree))
		valueList = append(valueList, int64(d.INodesUsed)-int64(base.INodesUsed))
		valueList = append(valueList, int64(d.UsedPercentage)-int64(base.UsedPercentage))
	}

	return helpers.WriteValueList(valueList, "")
}
