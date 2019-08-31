package stats

import (
	"log"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/process"
	"github.com/zkynetio/lynx/helpers"
)

type ProcessorDynamic struct {
	Processes      int
	PercentageUsed float64
	ValueList      []int64
}

func GetCPUByte() byte {
	cpuStat, err := process.Processes()
	helpers.PanicX(err)
	var psTotal float64
	// var count int
	for _, v := range cpuStat {
		ps, err := v.CPUPercent()
		if err != nil {
			log.Println("A process might disapear before we manage to stat it..")
			continue
		}
		// helpers.PanicX(err)
		psTotal = psTotal + ps
		// count++
	}
	psTotal = psTotal / float64(runtime.NumCPU())
	// log.Println("SAVING CPU:", psTotal, int8(psTotal), byte(int8(psTotal)))
	return byte(int8(psTotal))
}
func collectCPU(dp *DynamicPoint) {
	cpuStat, err := process.Processes()
	helpers.PanicX(err)
	dp.ProcessorDynamic = &ProcessorDynamic{}
	for _, v := range cpuStat {
		ps, err := v.CPUPercent()
		helpers.PanicX(err)
		dp.ProcessorDynamic.PercentageUsed = dp.ProcessorDynamic.PercentageUsed + ps
		dp.ProcessorDynamic.Processes++
	}
	dp.ProcessorDynamic.PercentageUsed = dp.ProcessorDynamic.PercentageUsed / float64(runtime.NumCPU())

}
func (p *ProcessorDynamic) GetFormattedBytes(basePoint bool) []byte {
	base := History.DynamicBasePoint.ProcessorDynamic
	if basePoint {
		base.ValueList = append(base.ValueList, int64(base.PercentageUsed))
		base.ValueList = append(base.ValueList, int64(base.Processes))
		return helpers.WriteValueList(base.ValueList, "")
	}

	prev := History.DynamicPreviousUpdatePoint.ProcessorDynamic
	p.ValueList = append(p.ValueList, int64(p.PercentageUsed))
	p.ValueList = append(p.ValueList, int64(p.Processes))
	return helpers.WriteValueList2(p.ValueList, base.ValueList, prev.ValueList, "")
}
func getProcesses() {
	ps, err := process.Processes()
	helpers.PanicX(err)
	//log.Println(ps)

	// for i, v := range netstuff {
	// }
	// select an invdividual process by id
	// pxs, err := process.NewProcess(6666)
	// helpers.PanicX(err)
	// log.Println(pxs.Status())

	var user float64
	for _, v := range ps {
		name, err := v.Name()

		helpers.PanicX(err)
		ps, err := v.CPUPercent()
		helpers.PanicX(err)

		stuff, err := v.Connections()
		helpers.PanicX(err)
		username, err := v.Username()
		helpers.PanicX(err)

		status, err := v.Status()
		helpers.PanicX(err)
		// // // connections, err := v.Connections()
		// // // helpers.PanicX(err)

		// terminal, err := v.Terminal()
		// helpers.PanicX(err)

		times, err := v.Times()
		helpers.PanicX(err)
		user = user + ps

		// run directory
		// typex, err := v.Cwd()
		// helpers.PanicX(err)
		// // typex, err := v.MemoryInfo()
		// helpers.PanicX(err)
		//if terminal != "" {

		if username != "root" {
			log.Println(len(stuff), v.Pid, username, status, name, ps, times.User, times.System, times.Idle, times.Iowait)
		}

		// for _, v := range connections {
		// 	if v.Status != "NONE" {
		// 		log.Println(v)
		// 	}
		// }
		//}

	}
	log.Println(user / float64(runtime.NumCPU()))

	os.Exit(1)
}
