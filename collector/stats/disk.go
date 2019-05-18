package stats

import (
	"bytes"
	"log"
	"strconv"
	"strings"

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

func (d *DiskDynamic) GetFormattedBytes() []byte {
	var data []byte
	var headers []byte
	var dataAndHeader []byte
	var buffer bytes.Buffer
	index := History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic
	// Disk free space state change
	if index.Free != d.Free {
		d := index.Free - d.Free
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}

	if index.Used != d.Used {
		d := index.Used - d.Used
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}

	if index.INodesTotal != d.INodesTotal {
		d := index.INodesTotal - d.INodesTotal
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	if index.INodesFree != d.INodesFree {
		d := index.INodesFree - d.INodesFree
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	if index.INodesUsed != d.INodesUsed {
		d := index.INodesUsed - d.INodesUsed
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	if index.UsedPercentage != d.UsedPercentage {
		d := index.UsedPercentage - d.UsedPercentage
		length := helpers.WriteIntToBuffer(&buffer, int64(d))
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	dataAndHeader = append(dataAndHeader, headers...)
	dataAndHeader = append(dataAndHeader, data...)
	log.Println("formatted bytes", dataAndHeader)
	return dataAndHeader
}
func (d *DiskDynamic) GetFormattedString() string {
	var diskSlice []string

	// TODO? ALERTS ON DISK CHANGES !!!!!
	diskSlice = append(diskSlice, strconv.FormatFloat(History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.UsedPercentage, 'f', 6, 64))

	if History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.Free != d.Free {
		diskSlice = append(diskSlice, strconv.Itoa(int(d.Free-History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.Free)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.Used != d.Used {
		diskSlice = append(diskSlice, strconv.Itoa(int(d.Used-History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.Used)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesTotal != d.INodesTotal {
		diskSlice = append(diskSlice, strconv.Itoa(int(d.INodesTotal-History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesTotal)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesFree != d.INodesFree {
		diskSlice = append(diskSlice, strconv.Itoa(int(d.INodesFree-History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesFree)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	if History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesUsed != d.INodesUsed {
		diskSlice = append(diskSlice, strconv.Itoa(int(d.INodesUsed-History.DynamicPointMap[HighestHistoryIndex-1].DiskDynamic.INodesUsed)))
	} else {
		diskSlice = append(diskSlice, "")
	}

	return strings.Join(diskSlice, ",")
}
