package stats

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	var valueList []int64
	// Disk free space state change
	if index.Free != d.Free {
		valueList = append(valueList, int64(index.Free)-int64(d.Free))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}

	if index.Used != d.Used {
		valueList = append(valueList, int64(index.Used)-int64(d.Used))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}

	if index.INodesTotal != d.INodesTotal {
		valueList = append(valueList, int64(index.INodesTotal)-int64(d.INodesTotal))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}
	if index.INodesFree != d.INodesFree {
		valueList = append(valueList, int64(index.INodesFree)-int64(d.INodesFree))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}
	if index.INodesUsed != d.INodesUsed {
		valueList = append(valueList, int64(index.INodesUsed)-int64(d.INodesUsed))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}
	if int64(index.UsedPercentage) != int64(d.UsedPercentage) {
		valueList = append(valueList, int64(index.UsedPercentage)-int64(d.UsedPercentage))
		// length := helpers.WriteIntToBuffer(&buffer, int64(d))
		// headers = append(headers, length)
		// data = append(data, buffer.Bytes()...)
		// buffer.Reset()
	}
	// var bs []
	// binary.LittleEndian.PutUint64(bs, uint64(1111))

	// x := binary.LittleEndian.Uint64(bs[0:9])
	// log.Println(x)
	buf := new(bytes.Buffer)
	var num int64 = 1234
	err := binary.Write(buf, binary.LittleEndian, num)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}

	// log.Println(int(buf.Bytes()))

	log.Println(valueList)
	for _, v := range valueList {
		length := helpers.WriteIntToBuffer(&buffer, v)
		headers = append(headers, length)
		data = append(data, buffer.Bytes()...)
		buffer.Reset()
	}
	log.Println()
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
