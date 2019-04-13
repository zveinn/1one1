package stats

import (
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
	"github.com/zkynetio/lynx/helpers"
)

var History *HistoryBuffer

const HighestHistoryIndex = 5

type DynamicPoint struct {
	MemoryDynamic
	LoadDynamic
	DiskDynamic
	CPU string
	//Host      string
	General   string
	Load1MIN  float64
	Load5MIN  float64
	Load15MIN float64
}

type StaticPoint struct {
}

type HistoryBuffer struct {
	StaticPointMap  map[int]*StaticPoint
	DynamicPointMap map[int]*DynamicPoint
}

func fetchDISK() {
	diskStat, err := disk.Usage("/")
	helpers.PanicX(err)
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
	helpers.PanicX(err)
	log.Println("CURRENT CPU PERCENTAGE:", percentage)
}
func fetchCPU() {
	cpuStat, err := cpu.Info()
	helpers.PanicX(err)
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
	helpers.PanicX(err)
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
	helpers.PanicX(err)
	//log.Println(ps)
	for _, v := range ps {
		usernme, err := v.Username()
		helpers.PanicX(err)

		status, err := v.Status()
		helpers.PanicX(err)
		connections, err := v.Connections()
		helpers.PanicX(err)
		log.Println(usernme, v.Pid, status)
		for _, v := range connections {
			if v.Status != "NONE" {
				log.Println(v)
			}
		}
	}
}

// ANAL-ISIS

func InitStats() {
	History = &HistoryBuffer{
		DynamicPointMap: make(map[int]*DynamicPoint),
		StaticPointMap:  make(map[int]*StaticPoint),
	}
}
func CollectDynamicData() string {
	// move DynamicPoints
	// TODO: do better
	if History.DynamicPointMap[HighestHistoryIndex] != nil {
		History.DynamicPointMap[HighestHistoryIndex-1] = History.DynamicPointMap[HighestHistoryIndex]
		//helpers.DebugLog("second index", History.DynamicPointMap[HighestHistoryIndex-1])
	} else {
		History.DynamicPointMap[HighestHistoryIndex-1] = &DynamicPoint{}
	}

	History.DynamicPointMap[HighestHistoryIndex] = &DynamicPoint{}
	//helpers.DebugLog("current index", History.DynamicPointMap[HighestHistoryIndex])
	//helpers.DebugLog("second index", History.DynamicPointMap[HighestHistoryIndex-1])

	// start
	collectLoad(History.DynamicPointMap[HighestHistoryIndex])
	collectMemory(History.DynamicPointMap[HighestHistoryIndex])
	collectDiskDynamic(History.DynamicPointMap[HighestHistoryIndex])

	return PrepareDataForShipping()
}

func AnalyzeData() {
	log.Println("ANALYZE data point")
}
func PrepareDataForShipping() (data string) {
	data = History.DynamicPointMap[HighestHistoryIndex].MemoryDynamic.GetFormattedString() + ":"
	data = data + History.DynamicPointMap[HighestHistoryIndex].LoadDynamic.GetFormattedString() + " :"
	data = data + History.DynamicPointMap[HighestHistoryIndex].DiskDynamic.GetFormattedString()

	return
}
