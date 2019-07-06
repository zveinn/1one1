package stats

import (
	"log"
	"os"
	"runtime"

	"github.com/shirou/gopsutil/process"
	"github.com/zkynetio/lynx/helpers"
)

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
