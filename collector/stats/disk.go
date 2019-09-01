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

func GetDiskByte() byte {
	diskStat, err := disk.Usage("/")
	helpers.PanicX(err)
	// log.Println("SAVING DISK:", diskStat.UsedPercent, int8(diskStat.UsedPercent), byte(int8(diskStat.UsedPercent)))
	return byte(int8(diskStat.UsedPercent))
}
