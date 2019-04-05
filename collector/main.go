package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	load "github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

var NS []string

// NOTES
// check for duplicate port/ip on collectors

func debugLog(v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(v...)
	}
}
func panicX(err error) {
	if err != nil {
		panic(err)
	}
}

type Collector struct {
	Buffer             chan string
	RecoveryFile       string
	TAG                string
	Controllers        []*Controller
	mutex              sync.Mutex
	MaintainerInterval int
	ListenerInterval   int
}

func (c *Collector) GetIntervalsFromEnvironmentVariables() {
	// set listener interval
	listenerInterval, err := strconv.Atoi(os.Getenv("LISTENER_INTERVAL"))
	if err != nil {
		debugLog("Error setting listener interval, default valur selected", err)
		listenerInterval = 5
	}
	c.ListenerInterval = listenerInterval

	// set maintainer interval
	maintainerInterval, err := strconv.Atoi(os.Getenv("MAINTAINER_INTERVAL"))
	if err != nil {
		debugLog("Error setting maintainer interval, default value selected", err)
		maintainerInterval = 5
	}
	c.MaintainerInterval = maintainerInterval
}
func (c *Collector) AddController(cont *Controller) {
	c.mutex.Lock()
	c.Controllers = append(c.Controllers, cont)
	c.mutex.Unlock()
}
func (c *Collector) CleanupOnExit() {
	for _, controller := range c.Controllers {
		log.Println("Closing:", controller.Address)
		if controller.Conn != nil {
			_ = controller.Conn.Close()
		}
	}
}

type Controller struct {
	Address     string
	Active      bool
	HasListener bool
	Conn        net.Conn
	Retry       int
	mutex       sync.Mutex
	NSDelivered bool
	Send        chan string
	//InactiveSince time.Time
}

func (c *Controller) ChangeActiveStatus(status bool) {
	c.mutex.Lock()
	c.Active = status
	c.mutex.Unlock()
}
func (c *Controller) HaveNamespacesBeenDelivered(delivered bool) {
	c.mutex.Lock()
	c.NSDelivered = delivered
	c.mutex.Unlock()
}
func (c *Controller) ChangeListenerStatus(status bool) {
	c.mutex.Lock()
	c.HasListener = status
	c.mutex.Unlock()
}
func (c *Controller) Setconnection(conn net.Conn) {
	c.mutex.Lock()
	c.Conn = conn
	c.mutex.Unlock()
}

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}
}

func (c *Controller) Listen() {
	defer func() {
		debugLog("defering read pipe from", c.Address)
		close(c.Send)
		c.ChangeActiveStatus(false)
		c.ChangeListenerStatus(false)
		c.Setconnection(nil)
		c.HaveNamespacesBeenDelivered(false)
	}()

	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')
		// TODO: handle better
		if err != nil || msg == "c\n" {
			_ = c.Conn.Close()
			debugLog("Error and/or message in read pipe from" + c.Address + " // " + msg + " //" + err.Error())
			break
		}

		if strings.Contains(msg, "ns:") {
			namespaces := strings.Split(strings.Split(strings.TrimSuffix(msg, "\n"), ":")[1], ",")
			log.Println("NAMESPACES:", namespaces)
			NS = namespaces
			// TODO: deliver base stats
			_, err := c.Conn.Write([]byte(getHost() + "\n"))
			panicX(err)
			c.HaveNamespacesBeenDelivered(true)
		}
		debugLog("IN:", msg)
		debugLog("NS:", NS)
		//go handleMessageFromController(msg)
	}

}

func (collector *Collector) engageControllerListeners() {
	for {
		time.Sleep(time.Duration(collector.ListenerInterval) * time.Second)
		for _, controller := range collector.Controllers {
			if !controller.Active || controller.HasListener {
				continue
			}
			debugLog("Engaging controller listener to", controller.Address)
			go controller.Listen()
			go controller.OpenSendChannel()
			controller.ChangeListenerStatus(true)
			_, err := controller.Conn.Write([]byte("k\n"))
			panicX(err)
		}
	}

}
func (collector *Collector) maintainControllerConnections() {
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		debugLog("5 second controller maintnance starting ...")
		for _, controller := range collector.Controllers {
			if controller.Active {
				continue
			}
			if err := dialAndHandshake(controller, collector.TAG); err != nil {
				debugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			debugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)
		}
	}

}

func dialController(controller *Controller) (err error) {
	conn, err := net.Dial("tcp", controller.Address)
	if err != nil {
		return
	}
	controller.Setconnection(conn)
	return
}
func handShakeWithController(controller *Controller, tag string) (err error) {
	_, err = controller.Conn.Write([]byte(tag + "\n"))
	panicX(err)
	message, err := bufio.NewReader(controller.Conn).ReadString('\n')
	log.Println(string(message))
	if err != nil || message != "k\n" {
		_ = controller.Conn.Close()
		controller.Setconnection(nil)
		// TODO: handle better
		err = errors.New("messsage from controller was:" + message + " // pipe read error was" + err.Error())
	}

	return
}
func dialAndHandshake(controller *Controller, tag string) (err error) {
	if err = dialController(controller); err != nil {
		return
	}
	if err = handShakeWithController(controller, tag); err != nil {
		return
	}
	return
}
func connectToControllers(controllers string, tag string, collector *Collector) {
	for _, v := range strings.Split(controllers, ",") {
		controller := &Controller{Address: v, Active: false, HasListener: false, NSDelivered: false}
		collector.AddController(controller)

		debugLog("Connecting to:", v)
		if err := dialAndHandshake(controller, tag); err != nil {
			debugLog("CONTROLLER COM. ERROR:", controller.Address)
			continue
		}
		debugLog("Connected to:", controller.Address)
		controller.ChangeActiveStatus(true)
	}
}

func main() {

	loadEnvironmentVariables()

	// Initialize a new controller
	collector := &Collector{
		TAG:          os.Getenv("TAG"),
		RecoveryFile: os.Getenv("RCOVERYFILE"),
	}
	collector.GetIntervalsFromEnvironmentVariables()
	defer collector.CleanupOnExit()

	connectToControllers(
		os.Getenv("CONTROLLERS"),
		os.Getenv("TAG"),
		collector,
	)

	go collector.engageControllerListeners()
	go collector.maintainControllerConnections()
	go collector.collectStats()

	if os.Getenv("DEBUG") == "true" {
		debugLog(collector)
		debugLog("Collector running...")
	}

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func (collector *Collector) collectStats() {
	defer func() {
		if r := recover(); r != nil {
			debugLog("recovered in stats collecting rutine", r)
			return
		}
	}()
	for {

		//fetchMemory()
		//fetchDISK()
		//fetchCPU()
		//fetchCPUPercentage()
		//getHost()
		//fetchNetworkIFS()
		//log.Println(getUploadDownload("wlp2s0"))
		//getEntropy()
		//getUsers()
		//getProcesses()
		//getLoad()
		//getActiveSessions()

		time.Sleep(100 * time.Millisecond)
		// TODO: make randomizer !
		for _, controller := range collector.Controllers {
			if controller.Active && controller.NSDelivered {
				controller.Send <- "boop\n"
			}
		}
	}

}

func (c *Controller) OpenSendChannel() {
	defer func() {
		log.Println("Closing send loop to controller", c.Address)
	}()
	c.Send = make(chan string, 10000)
	for {
		data, errx := <-c.Send
		if !errx {
			break
		}
		debugLog("sending to controller:", data)
		_, err := c.Conn.Write([]byte(data))
		if err != nil {
			debugLog("ERROR WHEN WRITING STATS:", err)
			close(c.Send)
			break
		}
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

func getHost() string {
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
