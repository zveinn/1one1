package stats

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
)

func GetCPUByte() byte {

	percent, _ := cpu.Percent(time.Second, true)
	var total float64
	for _, v := range percent {
		total = total + v
	}

	// cpuStat, err := process.Processes()
	// helpers.PanicX(err)
	// var psTotal float64
	// // var count int
	// for _, v := range cpuStat {
	// 	ps, err := v.CPUPercent()
	// 	if err != nil {
	// 		// log.Println("A process might disapear before we manage to stat it..")
	// 		continue
	// 	}
	// 	// helpers.PanicX(err)
	// 	psTotal = psTotal + ps
	// 	// count++
	// }
	// psTotal = psTotal / float64(runtime.NumCPU())
	// log.Println("SAVING CPU:", psTotal, int8(psTotal), byte(int8(psTotal)))
	return byte(int8(total / float64(runtime.NumCPU())))
}
