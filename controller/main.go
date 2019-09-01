package controller

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/lynx/helpers"
	"github.com/zkynetio/lynx/namespaces"
	"github.com/zkynetio/lynx/ui"
	"github.com/zkynetio/safelocker"
)

var GlobalController *Controller
var GlobalBrain *Brain

type Controller struct {
	safelocker.SafeLocker
	// Recovery in sqlite?
	//RecoveryFile string
	Config         ControllerConfig
	PORT           string
	IP             string
	Collectors     map[string]*Collector
	UIs            map[string]*ui.UI
	UIServer       *ui.UIServer
	UIParseChannel chan DPCollection
	UISendChannel  chan []byte
}
type Collector struct {
	TAG         string
	Conn        net.Conn
	LastCheckin time.Time
	SendChannel chan string
	Namespaces  map[int]string
}
type Brain struct {
	safelocker.SafeLocker
	Socket      net.Conn
	Address     string
	SendChannel chan []byte
	Alerting    []alerting.Alerting `json:"alerting"`
	Collecting  Collecting          `json:"collecting"`
}
type Collecting struct {
	Default []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"default"`
	Custom []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"custom"`
}
type ControllerConfig struct {
	Restart  bool
	IP       string
	Debug    bool
	Shutdown bool
	UI       struct {
		IP   string
		Port int
	}
	Collector struct {
		IP   string
		Port int
	}
}

func (b *Brain) connectToBrain(address string) (config ControllerConfig) {
	for {
		time.Sleep(1 * time.Second)
		var socket net.Conn
		socket, err := net.Dial("tcp", b.Address)
		if err != nil {
			log.Println("Error reaching out to the brain:", err)
			continue
		}

		lengthBytes := make([]byte, 3)
		_, err = socket.Read(lengthBytes)
		if err != nil {
			log.Println("Could not read control bytes from the brain:", err)
			continue
		}

		data := make([]byte, binary.LittleEndian.Uint16(lengthBytes[:2]))
		_, err = socket.Read(data)
		if err != nil {
			log.Println("Could not read data from the brain:", err)
			continue
		}
		if lengthBytes[2] == 0 {
			log.Println("The brain sent an error:", string(data))
			continue
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Println("Error parsing JSON from the brain:", string(data))
			continue
		}
		b.Socket = socket
		b.SendChannel = make(chan []byte, 10000)

		break
	}

	log.Println("/////////////////////////////////////////////////////")
	log.Println("I managed to reach the brain @", address)
	log.Println("Controller default IP:", config.IP)
	log.Println("Collector IP/PORT:", config.Collector)
	log.Println("UI IP/PORT:", config.UI)
	log.Println("Debug:", config.Debug)
	log.Println("/////////////////////////////////////////////////////")
	return
}
func Start(address string) {
	rand.Seed(time.Now().Unix())
	namespaces.Init()
	Brain := Brain{
		Address: address,
	}
	GlobalBrain = &Brain
	file, err := os.OpenFile("connectTo", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println("could not create file (connectTo) in root directory...")
		os.Exit(1)
	}

	reader := bufio.NewReader(file)
	line, _, _ := reader.ReadLine()

	if len(line) > 0 {
		Brain.Address = string(line)
		log.Println("Got the brains address from the connectTo file:", Brain.Address)
	} else {
		_, err = file.WriteAt([]byte(address), 0)
		if err != nil {
			log.Println("could not write connectTo flag to file (connectTo) in root directory")
			os.Exit(1)
		}
	}
	log.Println(address)
	config := Brain.connectToBrain(address)

	UIServer := ui.NewUIServer()
	UIServer.ClientList = make(map[string]*ui.UI)

	if config.UI.IP == "" {
		UIServer.IP = config.IP
	} else {
		UIServer.IP = config.UI.IP
	}

	UIServer.Port = strconv.Itoa(config.UI.Port)

	controller := Controller{
		Config:     config,
		Collectors: make(map[string]*Collector),
		UIServer:   UIServer,
	}
	// os.Exit(1)
	watcherChannel := make(chan int, 10)
	closeChannel := make(chan bool, 10)
	go Brain.MaintainLinkToBrain(watcherChannel, closeChannel)
	controller.UISendChannel = make(chan []byte, 1000000)
	controller.UIParseChannel = make(chan DPCollection, 100000)
	go ShipToUIS(controller.UISendChannel)
	go SaveToUIBuffer(controller.UIParseChannel, controller.UISendChannel)

	GlobalController = &controller

	defer controller.CleanupOnExit()
	go UIServer.Start(watcherChannel)
	go controller.start(watcherChannel, closeChannel)
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case index := <-watcherChannel:
			time.Sleep(1 * time.Second)
			if index == 1 {
				go UIServer.Start(watcherChannel)
			} else if index == 2 {
				go controller.start(watcherChannel, closeChannel)
			} else if index == 3 {
				GlobalController.Config = Brain.connectToBrain(address)
				go Brain.MaintainLinkToBrain(watcherChannel, closeChannel)
			}

			log.Println("Goroutine watcher detected a crash in goroutine number:", index)
			break
		case signal := <-stop:
			log.Println("Signal received (", signal, ") controller will now exit gracefully")
			controller.CleanupOnExit()

			os.Exit(1)
		}
	}

}

func (b *Brain) MaintainLinkToBrain(watcherChannel chan int, closeChannel chan bool) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("Brain link crahed !!", r, string(debug.Stack()))
		}
		watcherChannel <- 3
	}(watcherChannel)

	for {
		log.Println("listening to brain ...")
		controlBytes := make([]byte, 3)
		_, err := b.Socket.Read(controlBytes)
		if err != nil {
			log.Println("Error reading length bytes from brain", err)
			return
		}

		data := make([]byte, binary.LittleEndian.Uint16(controlBytes[:2]))
		_, err = b.Socket.Read(data)
		if err != nil {
			log.Println("Error reading data fom brain ...", err)
		}

		if controlBytes[2] == 1 {
			// config
			err = b.DecodeConfig(closeChannel, data)
			if err == nil {
				continue
			}
		} else if controlBytes[2] == 2 {
			// brain
			err = b.DecodeBrain(data)
			if err == nil {
				continue
			}
		}

	}
}
func RestartServers(closeChannel chan bool) {
	rand.Seed(time.Now().UnixNano())
	n := rand.Intn(5)
	time.Sleep(time.Duration(n) * time.Second)
	ui.UIHTTPsrv.Close()
	GlobalController.Lock()
	for _, v := range GlobalController.UIServer.ClientList {
		if v != nil {
			_ = v.Conn.Close()
		}
	}
	for _, v := range GlobalController.Collectors {
		if v != nil {
			_ = v.Conn.Close()
		}
	}
	GlobalController.Unlock()
	closeChannel <- true
}
func (b *Brain) DecodeBrain(data []byte) error {
	var brain Brain
	// log.Println(string(data))
	err := json.Unmarshal(data, &brain)
	if err != nil {
		log.Println("could not decode brain", err)
		return err
	}
	b.Lock()
	if len(brain.Alerting) > 0 {
		b.Alerting = brain.Alerting
	}
	if len(brain.Collecting.Custom) > 0 || len(brain.Collecting.Default) > 0 {
		b.Collecting = brain.Collecting
	}

	b.Unlock()
	log.Println(b.Collecting)
	log.Println(b.Alerting)
	return nil
}
func (b *Brain) DecodeConfig(closeChannel chan bool, data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("crashed while restarting..", r, string(debug.Stack()))
			GlobalController.SafeUnlock()
		}
	}()

	var config ControllerConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		log.Println(err)
		return err
	} else if config.IP == "" {
		return errors.New("Not updating config, IP missing")
	} else if config.Shutdown {
		log.Println("Brain told me to exit...")
		os.Exit(1)
	} else if config.Restart {
		RestartServers(closeChannel)
	}

	GlobalController.Lock()
	GlobalController.Config = config
	GlobalController.Unlock()
	return nil

}

func (c *Controller) CleanupOnExit() {
	log.Println("Post exit cleanup ..")
	for _, collector := range c.Collectors {
		if collector.Conn != nil {
			log.Println("Closing connection to:", collector.TAG, "@", collector.Conn.RemoteAddr().String())
			_ = collector.Conn.Close()
		}
	}
}

func (c *Controller) AddCollector(TAG string, collector *Collector) error {
	c.Lock()
	defer c.Unlock()
	if c.Collectors[TAG] != nil {
		log.Println("A controller already exists with this tag:", TAG)
		return errors.New("This TAG already exists")
	}
	c.Collectors[TAG] = collector
	return nil
}

func (c *Controller) RemoveCollector(TAG string) {
	c.Lock()
	c.Collectors[TAG] = nil
	c.Unlock()
}

func (controller *Controller) start(watcherChannel chan int, closeChannel chan bool) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("controller crahed !!", r, string(debug.Stack()))
		}
		watcherChannel <- 2
	}(watcherChannel)
	log.Println("listening on:", controller.Config.Collector.IP+":"+strconv.Itoa(controller.Config.Collector.Port))

	address := ""
	if controller.Config.Collector.IP == "" {
		address = address + controller.Config.IP + ":" + strconv.Itoa(controller.Config.Collector.Port)
	} else {
		address = address + controller.Config.Collector.IP + ":" + strconv.Itoa(controller.Config.Collector.Port)
	}

	ln, err := net.Listen("tcp", address)

	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		log.Println("Checking close channel ...")
		select {
		case <-closeChannel:
			log.Println("closing collector port..")
			_ = ln.Close()
			_ = conn.Close()
			return
		default:
			log.Println("nothing to close ... ")
		}
		helpers.PanicX(err)
		go receiveConnection(conn, controller)

	}
}

func receiveConnection(conn net.Conn, controller *Controller) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("recovered in connection receiver ", r, string(debug.Stack()))
		}
	}()
	log.Println("Connection from:", conn.RemoteAddr())

	message, _ := bufio.NewReader(conn).ReadString('\n')

	// STEP 2
	// Connect a collector
	connectCollector(&Collector{
		LastCheckin: time.Now(),
		Conn:        conn,
	}, controller, message)
}
func findCollectorNamespaces(collector *Collector) (namespaces []string) {
	var defaultNamespaces []string
	for _, v := range GlobalBrain.Collecting.Default {
		if strings.Contains(collector.TAG, v.Tag) || v.Tag == "*" {
			defaultNamespaces = append(defaultNamespaces, v.Namespaces...)
		}
	}
	var customNamespaces []string
	for _, v := range GlobalBrain.Collecting.Custom {
		if strings.Contains(collector.TAG, v.Tag) {
			customNamespaces = append(customNamespaces, v.Namespaces...)
		}
	}

	if len(customNamespaces) > 0 {
		log.Println("USING CUSTOM INDEXES:", customNamespaces)
		namespaces = customNamespaces
	} else {
		log.Println("USING DEFAULT INDEXES:", defaultNamespaces)
		namespaces = defaultNamespaces
	}
	return
}
func connectCollector(collector *Collector, controller *Controller, message string) {
	collector.TAG = strings.TrimSuffix(message, "\n")
	ns := findCollectorNamespaces(collector)
	collector.Namespaces = namespaces.MakeMapFromNamespaces(ns)

	defer func() {
		log.Println("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	log.Println("COLLECTOR:", collector.TAG)
	err := controller.AddCollector(collector.TAG, collector)
	if err != nil {
		_, _ = collector.Conn.Write([]byte("E: this TAG already exists\n"))
		_ = collector.Conn.Close()
		return
	}

	jsonIndexes, err := json.Marshal(ns)
	if err != nil {
		panic(err)
	}
	log.Println("SENDING INDEXES:", string(jsonIndexes))
	_, err = collector.Conn.Write([]byte(string(jsonIndexes) + "\n"))
	if err != nil {
		log.Println("could not establish coms with collector", err)
		controller.RemoveCollector(collector.TAG)
		return
	}

	log.Println("Starting general collection from:", collector.TAG)
	readFromConnectionOriginal(collector, controller)
}
func readFromConnectionOriginal(collector *Collector, controller *Controller) {
	reader := bufio.NewReader(collector.Conn)
	// TODO: move all parsing into go routine ?
	for {
		controlBytes := make([]byte, 3)
		_, err := reader.Read(controlBytes)
		if err != nil {
			log.Println("Closing Collector read pipe 1:", err)
			collector.Conn.Close()
			return
		}
		length := binary.LittleEndian.Uint16(controlBytes[0:2])
		// log.Println("DATA LENGTH BYTES:", controlBytes[0:2], "length:", length)
		data := make([]byte, length+8)
		_, err = reader.Read(data)
		if err != nil {
			log.Println("Closing Collector read pipe 2:", err)
			collector.Conn.Close()
			return
		}

		go controller.HandleDataPoint(collector, data, int(controlBytes[2]))

	}

}

func (c *Controller) HandleDataPoint(collector *Collector, data []byte, controlByte int) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("panic!", r, string(debug.Stack()))
		}
	}()
	var DPC DPCollection
	if controlByte == 4 {
		// log.Println("TIME:", timestamp)
		// log.Println("FULL DATA:", data)
		// log.Println("TAG:", tag, " CONTROL BYTE:", controlByte)
		DPC.Timestamp = binary.LittleEndian.Uint64(data[:8])
		DPC.Tag = collector.TAG
		DPC.ControlByte = controlByte
		DPC = ParseMinimumDataPoint(data[8:], collector.Namespaces)

		d := ""
		for _, v := range DPC.DPS {
			d = d + strconv.Itoa(v.Index) + "/" + strconv.Itoa(v.Value) + "  - "
		}
		log.Println(d)

	}

	c.UIParseChannel <- DPC
}
