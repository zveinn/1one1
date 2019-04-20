package stats

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/zkynetio/lynx/helpers"
)

var History *HistoryBuffer

const HighestHistoryIndex = 5

type DynamicPoint struct {
	MemoryDynamic
	LoadDynamic
	DiskDynamic
	NetworkDynamic
	EntropyDynamic
	CPU string
	//Host      string
	General   string
	Load1MIN  float64
	Load5MIN  float64
	Load15MIN float64
}

type StaticPoint struct {
	NetworkStaticList map[string]*NetworkStatic
	HostStatic
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
	collectNetworkDownloadAndUpload(History.DynamicPointMap[HighestHistoryIndex])
	collectLoad(History.DynamicPointMap[HighestHistoryIndex])
	collectMemory(History.DynamicPointMap[HighestHistoryIndex])
	collectDiskDynamic(History.DynamicPointMap[HighestHistoryIndex])
	collectEntropy(History.DynamicPointMap[HighestHistoryIndex])

	return PrepareDataForShipping()
}

func AnalyzeData() {
	log.Println("ANALYZE data point")
}
func PrepareDataForShipping() (data string) {
	getProcesses()
	data = History.DynamicPointMap[HighestHistoryIndex].MemoryDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].LoadDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].DiskDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].EntropyDynamic.GetFormattedString() + ";"
	data = data + History.DynamicPointMap[HighestHistoryIndex].NetworkDynamic.GetFormattedString()
	return
}

func GetStaticDataPoint() string {
	data := "h:::"
	data = data + GetHost()
	log.Println(data)
	return data
}
