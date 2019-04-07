package stats

import (
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

func panicX(err error) {
	if err != nil {
		panic(err)
	}
}
func STATSMemory() string {
	vmStat, err := mem.VirtualMemory()
	panicX(err)
	log.Println("=========== MEMORY =================")
	log.Println("Total:", vmStat.Total)
	log.Println("Used:", vmStat.Used)
	log.Println("Free:", vmStat.Free)
	log.Println("Shared:", vmStat.Shared)
	log.Println("Buff/cached:", vmStat.Cached+vmStat.Buffers)
	log.Println("Available:", vmStat.Available)
	log.Println("swap")
	log.Println("Swap Free:", vmStat.SwapFree)
	log.Println("Swap Cached:", vmStat.SwapCached)
	log.Println("Swap Total:", vmStat.SwapTotal)
	log.Println("=========== MEMORY =================")
	return vmStat.String()
}

func fetchDISK() {
	diskStat, err := disk.Usage("/")
	panicX(err)
	log.Println(disk.GetDiskSerialNumber("/dev/sda2"))
	log.Println("=========== DISK / =================")
	log.Println("Total: ", diskStat.Total)
	log.Println("Free: ", diskStat.Free)
	log.Println("Used: ", diskStat.Used)
	log.Println("UsedPercent: ", diskStat.UsedPercent)
	log.Println("Path: ", diskStat.Path)
	log.Println("Fstype: ", diskStat.Fstype)

	log.Println("InodesTotal: ", diskStat.InodesTotal)
	log.Println("InodesUsed: ", diskStat.InodesUsed)
	log.Println("InodesFree: ", diskStat.InodesFree)
	log.Println("=========== DISK =================")
}

func fetchCPUPercentage() {
	percentage, err := cpu.Percent(0, true)
	panicX(err)
	log.Println("CURRENT CPU PERCENTAGE:", percentage)
}
func fetchCPU() {
	cpuStat, err := cpu.Info()
	panicX(err)
	for _, v := range cpuStat {
		log.Println("=========== CPu=================")
		//log.Println("PERCENTAGE:", percentage[i])
		log.Println("CPU:", v.CPU)
		log.Println("VendorID:", v.VendorID)
		log.Println("Cores:", v.Cores)
		log.Println("CoreID:", v.CoreID)
		log.Println("MODEL", v.Model)
		log.Println("ModelName:", v.ModelName)
		log.Println("Microcode:", v.Microcode)
		log.Println("PhysicalID:", v.PhysicalID)
		log.Println("Mhz:", v.Mhz)
		log.Println("Flags:", v.Flags)
		log.Println("Model:", v.Model)
		log.Println("CacheSize:", v.CacheSize)
		log.Println("Stepping:", v.Stepping)

		log.Println("============================")
	}

}

func GetHost() string {
	hostStat, err := host.Info()
	panicX(err)
	host := "h:::" + hostStat.Hostname
	host = host + "," + hostStat.HostID
	host = host + "," + hostStat.KernelVersion
	host = host + "," + hostStat.OS
	host = host + "," + hostStat.Platform
	host = host + "," + hostStat.PlatformFamily
	host = host + "," + hostStat.PlatformVersion
	//	host = host + "," + strconv.FormatUint(hostStat.Procs, 10)
	host = host + "," + strconv.FormatUint(hostStat.Uptime, 10)
	host = host + "," + hostStat.VirtualizationRole
	host = host + "," + hostStat.VirtualizationSystem
	host = host + "," + strconv.FormatUint(hostStat.BootTime, 10)
	return host
}

func fetchNetworkIFS() {
	interfStat, err := net.Interfaces()
	panicX(err)

	for _, interf := range interfStat {

		log.Println("Name:", interf.Name)
		log.Println("HardwareAddr:", interf.HardwareAddr)
		addrs, err := interf.Addrs()
		panicX(err)
		for _, addr := range addrs {
			log.Println("ADDR:::", addr)
		}
		maddrs, err := interf.MulticastAddrs()
		panicX(err)
		for _, addr := range maddrs {
			log.Println("MULTICAST ADDS:::", addr)
		}
		log.Println("flags", interf.Flags)
		log.Println("mtu", interf.MTU)
		log.Println("index", interf.Index)
		//for _, flag := range interf.Flags {
		//	log.Println("FLAG:::", flag)
		//}
		getUploadDownload(interf.Name)

	}
}

func getUploadDownload(ifname string) {
	file, err := ioutil.ReadFile("/proc/net/dev") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	//defer file.Close()
	filestring := string(file)
	//log.Println("=============================")
	//log.Println(filestring)
	//log.Println("=============================")
	fileSplit := strings.Split(filestring, "\n")
	for _, v := range fileSplit {
		if strings.Contains(v, ifname) {
			vSplit := strings.Split(v, " ")
			//for i, v := range vSplit {
			//	log.Println(i, ":", v)
			//}
			log.Println("Download bytes:", vSplit[1])
			log.Println("Download packets:", vSplit[2])
			log.Println("Upload bytes:", vSplit[39])
			log.Println("Upload packates:", vSplit[41])
		}
		//if v == ifname {
		//	log.Println(i, ":", v)
		//	return "FOUND IT!"
		//}
	}
	//log.Println(fileSplit[247])

}

func STATSEntropy() string {
	file, err := ioutil.ReadFile("/proc/sys/kernel/random/entropy_avail") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	log.Println("entropy", string(file))
	return string(file)
}

func getUsers() {
	file, err := ioutil.ReadFile("/etc/passwd") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	fileSplit := strings.Split(string(file), "\n")
	log.Println(fileSplit)
	//for _, v := range fileSplit {
	//	username := strings.Split(v, ":")[0]
	//
	//	if username != "" {
	//		out, err := exec.Command("bash", "-c", "ps -u "+username).Output()
	//		if err != nil {
	//			//log.Println(err)
	//		}
	//		split := strings.Split(string(out), "\n")
	//		if len(split) > 2 {
	//			log.Println("USERNAME:", username)
	//			log.Println(string(out))
	//		}
	//
	//	}
	//
	//}

	//log.Println("users", string(file))
}

func getActiveSessions() {

	out, err := exec.Command("bash", "-c", "who -a").Output()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(out)
}

func getProcesses() {
	ps, err := process.Processes()
	panicX(err)
	//log.Println(ps)
	for _, v := range ps {
		usernme, err := v.Username()
		panicX(err)

		status, err := v.Status()
		panicX(err)
		connections, err := v.Connections()
		panicX(err)
		log.Println(usernme, v.Pid, status)
		for _, v := range connections {
			if v.Status != "NONE" {
				log.Println(v)
			}
		}
	}
}

func getLoad() {
	ld, err := load.Avg()
	panicX(err)
	log.Println(ld.String())
}
