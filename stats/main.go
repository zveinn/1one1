package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

func dealwithErr(err error) {
	if err != nil {
		fmt.Println(err)
		//os.Exit(-1)
	}
}

func GetHardwareData(w http.ResponseWriter, r *http.Request) {
	runtimeOS := runtime.GOOS
	// memory
	vmStat, err := mem.VirtualMemory()
	dealwithErr(err)

	// disk - start from "/" mount point for Linux
	// might have to change for Windows!!
	// don't have a Window to test this out, if detect OS == windows
	// then use "\" instead of "/"

	diskStat, err := disk.Usage("/")
	dealwithErr(err)

	// cpu - get CPU number of cores and speed
	cpuStat, err := cpu.Info()
	dealwithErr(err)
	percentage, err := cpu.Percent(0, true)
	dealwithErr(err)

	// host or machine kernel, uptime, platform Info
	hostStat, err := host.Info()
	dealwithErr(err)

	// get interfaces MAC/hardware address
	interfStat, err := net.Interfaces()
	dealwithErr(err)

	html := runtimeOS + "<br>"
	html = html + strconv.FormatUint(vmStat.Total, 10) + "  <br>"
	html = html + strconv.FormatUint(vmStat.Free, 10) + " <br>"
	html = html + strconv.FormatFloat(vmStat.UsedPercent, 'f', 2, 64) + "%<br>"

	// get disk serial number.... strange... not available from disk package at compile time
	// undefined: disk.GetDiskSerialNumber
	//serial := disk.GetDiskSerialNumber("/dev/sda")

	//html = html + "Disk serial number: " + serial + "<br>"

	html = html + strconv.FormatUint(diskStat.Total, 10) + "  <br>"
	html = html + strconv.FormatUint(diskStat.Used, 10) + " <br>"
	html = html + strconv.FormatUint(diskStat.Free, 10) + " <br>"
	html = html + strconv.FormatFloat(diskStat.UsedPercent, 'f', 2, 64) + "%<br>"

	// since my machine has one CPU, I'll use the 0 index
	// if your machine has more than 1 CPU, use the correct index
	// to get the proper data
	html = html + strconv.FormatInt(int64(cpuStat[0].CPU), 10) + "<br>"
	html = html + cpuStat[0].VendorID + "<br>"
	html = html + cpuStat[0].Family + "<br>"
	html = html + strconv.FormatInt(int64(cpuStat[0].Cores), 10) + "<br>"
	html = html + cpuStat[0].ModelName + "<br>"
	html = html + strconv.FormatFloat(cpuStat[0].Mhz, 'f', 2, 64) + " MHz <br>"

	for idx, cpupercent := range percentage {
		html = html + strconv.Itoa(idx) + "] " + strconv.FormatFloat(cpupercent, 'f', 2, 64) + "%<br>"
	}

	html = html + hostStat.Hostname + "<br>"
	html = html + strconv.FormatUint(hostStat.Uptime, 10) + "<br>"
	html = html + strconv.FormatUint(hostStat.Procs, 10) + "<br>"

	// another way to get the operating system name
	// both darwin for Mac OSX, For Linux, can be ubuntu as platform
	// and linux for OS

	html = html + hostStat.OS + "<br>"
	html = html + hostStat.Platform + "<br>"

	// the unique hardware id for this machine
	html = html + hostStat.HostID + "<br>"

	for _, interf := range interfStat {

		html = html + interf.Name + "<br>"

		if interf.HardwareAddr != "" {
			html = html + interf.HardwareAddr + "<br>"
		}

		for _, flag := range interf.Flags {
			html = html + flag + "<br>"
		}

		for _, addr := range interf.Addrs {
			html = html + addr.String() + "<br>"

		}

	}

	w.Write([]byte(html))

}

func SayName(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, I'm a machine and my name is [whatever]"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", SayName)
	mux.HandleFunc("/gethwdata", GetHardwareData)

	http.ListenAndServe(":8080", mux)

}
