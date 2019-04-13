package stats

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/disk"
	"github.com/zkynetio/lynx/helpers"
)

type Disk struct {
	Serial         string
	Total          uint64
	Free           uint64
	Used           uint64
	UsedPercentage float64
	Path           string
	FSType         string
	INodesTotal    uint64
	INodesUsed     uint64
	INodesFree     uint64
}

func collectDisk(dp *DataPoint) {
	diskStat, err := disk.Usage("/")
	helpers.PanicX(err)

	dp.Disk = Disk{
		Total:          diskStat.Total,
		Free:           diskStat.Free,
		Used:           diskStat.Used,
		UsedPercentage: diskStat.UsedPercent,
		Path:           diskStat.Path,
		FSType:         diskStat.Fstype,
		INodesFree:     diskStat.InodesFree,
		INodesTotal:    diskStat.InodesTotal,
		INodesUsed:     diskStat.InodesUsed,
	}
}

func (d *Disk) GetFormattedString() string {
	var diskSlice []string

	// TODO? ALERTS ON DISK CHANGES !!!!!
	diskSlice = append(diskSlice, strconv.FormatFloat(History.DataPointMap[HighestHistoryIndex-1].Disk.UsedPercentage, 'f', 6, 64))

	if History.DataPointMap[HighestHistoryIndex-1].Disk.Free != d.Free {
		diskSlice = append(diskSlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Disk.Free-d.Free)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Disk.Used != d.Used {
		diskSlice = append(diskSlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Disk.Used-d.Used)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Disk.INodesTotal != d.INodesTotal {
		diskSlice = append(diskSlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Disk.INodesTotal-d.INodesTotal)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Disk.INodesFree != d.INodesFree {
		diskSlice = append(diskSlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Disk.INodesFree-d.INodesFree)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DataPointMap[HighestHistoryIndex-1].Disk.INodesUsed != d.INodesUsed {
		diskSlice = append(diskSlice, strconv.Itoa(int(History.DataPointMap[HighestHistoryIndex-1].Disk.INodesUsed-d.INodesUsed)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	return strings.Join(diskSlice, ",")
}
