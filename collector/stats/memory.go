package stats

import (
	"github.com/shirou/gopsutil/mem"
	"github.com/zkynetio/lynx/helpers"
)

func GetMemoryByte() byte {
	vmStat, err := mem.VirtualMemory()
	helpers.PanicX(err)
	// log.Println("SAVING MEMORY:", vmStat.UsedPercent, int8(vmStat.UsedPercent), byte(int8(vmStat.UsedPercent)))
	return byte(int8(vmStat.UsedPercent))
}
