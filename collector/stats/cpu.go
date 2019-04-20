package stats

import (
	"log"
	"os"

	"github.com/shirou/gopsutil/process"
	"github.com/zkynetio/lynx/helpers"
)

func getProcesses() {
	ps, err := process.Processes()
	helpers.PanicX(err)
	//log.Println(ps)
	for _, v := range ps {
		name, err := v.Name()
		helpers.PanicX(err)

		usernme, err := v.Username()
		helpers.PanicX(err)

		status, err := v.Status()
		helpers.PanicX(err)
		connections, err := v.Connections()
		helpers.PanicX(err)

		terminal, err := v.Terminal()
		helpers.PanicX(err)

		times, err := v.Times()
		helpers.PanicX(err)

		//if terminal != "" {
		log.Println(name, usernme, v.Pid, status, terminal, times)
		for _, v := range connections {
			if v.Status != "NONE" {
				log.Println(v)
			}
		}
		//}

	}

	os.Exit(1)
}
