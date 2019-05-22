package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/zkynetio/lynx/helpers"
)

func main() {
	rand.Seed(time.Now().Unix())
	loadEnvironmentVariables()

	controller := Controller{
		IP:                   os.Getenv("IP"),
		PORT:                 os.Getenv("PORT"),
		Collectors:           make(map[string]*Collector),
		Buffer:               make(chan *DataPoint, 10000),
		BufferDirectoryPath:  "./buffers/",
		MinLinesInBufferFile: 10,
		CollectionBuffer:     make(map[string][]*DataPoint),
	}

	LiveBuffer = &Collection{
		Map: make(map[string]map[string]map[string][][]byte),
	}
	defer controller.CleanupOnExit()
	go controller.EngageBufferPipe()
	controller.start()

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

type UI struct {
	Conn        net.Conn
	Filter      string
	SendChannel chan string
}
type Stat struct {
	Label       string
	Value       int
	StringValue string
}

type Controller struct {
	Buffer chan *DataPoint
	// Recovery in sqlite?
	//RecoveryFile string
	BufferDirectoryPath  string
	PORT                 string
	IP                   string
	Collectors           map[string]*Collector
	UIs                  map[string]*UI
	mutex                sync.Mutex
	MinLinesInBufferFile int
	CollectionBuffer     map[string][]*DataPoint
	mux                  sync.Mutex
}
type Collection struct {
	// instace,namespace,[year.month.day.hour.minute.second],[]byte
	Map map[string]map[string]map[string][][]byte
	Mux sync.Mutex
}

var LiveBuffer *Collection

// mat zoe? abive and beyond
func (c *Controller) CleanupOnExit() {
	helpers.DebugLog("Cleaning up on exit...")
	for _, collector := range c.Collectors {
		if collector.Conn != nil {
			_, _ = collector.Conn.Write([]byte("c\n"))
			helpers.DebugLog("Closing:", collector.TAG)
			_ = collector.Conn.Close()
		}
	}
}

type Collector struct {
	TAG         string
	Conn        net.Conn
	LastCheckin time.Time
	SendChannel chan string
}

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		helpers.DebugLog(err)
		log.Fatal("Error loading .env file")
	}
}
func (c *Controller) AddCollector(TAG string, collector *Collector) {
	c.mutex.Lock()
	c.Collectors[TAG] = collector
	c.mutex.Unlock()
}

func (c *Controller) RemoveCollector(TAG string) {
	c.mutex.Lock()
	c.Collectors[TAG] = nil
	c.mutex.Unlock()
}
func (c *Controller) AddUI(TAG string, ui *UI) {
	c.mutex.Lock()
	c.UIs[TAG] = ui
	c.mutex.Unlock()
}

func (c *Controller) RemoveUI(TAG string) {
	c.mutex.Lock()
	c.UIs[TAG] = nil
	c.mutex.Unlock()
}

func (controller *Controller) start() {

	helpers.DebugLog("listening on:", controller.IP+":"+controller.PORT)
	ln, err := net.Listen("tcp", controller.IP+":"+controller.PORT)
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		helpers.PanicX(err)
		go receiveConnection(conn, controller)
	}
}

func connectUI(ui *UI, controller *Controller, tag string) {
	helpers.DebugLog("UI:")
	controller.AddUI(tag+"TODORANDOMINT?", ui)
	for {
		outgoing := <-ui.SendChannel
		log.Println(outgoing)
		_, err := ui.Conn.Write([]byte(outgoing))
		//TODO: handle better
		helpers.PanicX(err)
	}
}

func deliverNamespaces(conn net.Conn) {
	_, err := conn.Write([]byte("NS|general.entropy:5,memory.total:5\n"))
	helpers.PanicX(err)
}
func receiveConnection(conn net.Conn, controller *Controller) {
	helpers.DebugLog("Connection from:", conn.RemoteAddr())

	message, _ := bufio.NewReader(conn).ReadString('\n')
	if message == "ui\n" {
		connectUI(&UI{Conn: conn}, controller, message)
	} else {
		// STEP 2
		// Connect a collector
		connectCollector(&Collector{
			LastCheckin: time.Now(),
			Conn:        conn,
		}, controller, message)
	}
}
func connectCollector(collector *Collector, controller *Controller, message string) {
	helpers.DebugLog("COLLECTOR:", strings.TrimSuffix(message, "\n"))
	collector.TAG = strings.TrimSuffix(message, "\n")

	controller.AddCollector(collector.TAG, collector)
	defer func() {
		helpers.DebugLog("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	// STEP 3
	// Send namespaces to collector
	deliverNamespaces(collector.Conn)

	// STEP 6
	// Listen for an OK
	msg, _ := bufio.NewReader(collector.Conn).ReadString('\n')
	if msg == "k\n" {
		helpers.DebugLog("K from collector")
	} else {
		helpers.DebugLog(msg)
		helpers.PanicX(errors.New("NOT OK FROM COLLECTOR"))
	}

	// STEP 7
	// Listen for host data
	msg, _ = bufio.NewReader(collector.Conn).ReadString('\n')
	if !strings.Contains(string(msg), "H||") {
		helpers.PanicX(errors.New("no host data found in handskae" + string(msg)))
	}
	helpers.DebugLog("HOST DATA:", string(msg))

	// STEP 9
	// Listen for final OK
	helpers.DebugLog("waiting for final ok ...")
	msg, _ = bufio.NewReader(collector.Conn).ReadString('\n')
	if msg != "k\n" {
		helpers.DebugLog(msg)
		helpers.PanicX(errors.New("NOT OK FROM COLLECTOR"))
	}

	log.Println("Starting general collection from:", collector.TAG)
	readFromConnectionOriginal(collector, controller)
}
func readFromConnectionOriginal(collector *Collector, controller *Controller) {
	reader := bufio.NewReader(collector.Conn)
	for {
		controlBytes := make([]byte, 3)
		_, err := reader.Read(controlBytes)
		if err != nil {
			panic(err)
		}
		// log.Println(lengthByte)
		length := binary.LittleEndian.Uint16(controlBytes[0:3])
		log.Println("Length:", length)
		log.Println("control byte:", controlBytes[2])
		// timestamp := binary.LittleEndian.Uint64(controlBytes[3:])

		// log.Println("timestamp", timestamp, "MS:", (timestamp / 1000000), controlBytes[3:])
		data := make([]byte, length)
		_, err = reader.Read(data)
		if err != nil {
			panic(err)
		}
		// log.Println(data)
		// ParseDataPoint(data)

		go controller.parseIncomingData(collector.TAG, data, int(controlBytes[2]))
	}

}

func (c *Controller) sendToAllUis(msg string) {
	for _, ui := range c.UIs {
		ui.SendChannel <- msg
	}
}

type DataPoint struct {
	Value       []byte
	Tag         string
	Timestamp   int64
	ControlByte int
}

func (c *Controller) parseIncomingData(tag string, data []byte, controlByte int) {

	//helpers.DebugLog("DATA:", msg)
	c.Buffer <- &DataPoint{
		Value:       data,
		Tag:         tag,
		ControlByte: controlByte,
	}
}

func (c *Controller) EngageBufferPipe() {

	// TODO: grow properly, we are writing strings.. but the buffer grows in bytes
	//	buffer.Grow(c.MinLinesInBufferFile * 500)
	// size := 0
	for {
		if len(c.Buffer) > 100 {
			//buffer.Grow(buffer.Len() * 2)
			helpers.DebugLog("chan length", len(c.Buffer))
		}

		dp := <-c.Buffer
		// CollectionBuffer.mux.Lock()
		// c.CollectionBuffer[dp.Tag] = append(c.CollectionBuffer[dp.Tag], dp)
		// CollectionBuffer.mux.Unlock()
		c.ParseDataPointIntoMemoryMap(dp)
		// log.Println(c.CollectionBuffer)
		// size = size + len(dp.Value)
		// if size > 3000 {
		// 	// write data
		// 	// go c.WriteBufferToFile()
		// }

	}
}
func (c *Controller) ParseDataPointIntoMemoryMap(dp *DataPoint) {

	// timestamp := dp.Timestamp.Unix()
	// tag := dp.Tag
	log.Println(dp.Value)
	timestamp := binary.LittleEndian.Uint64(dp.Value[:8])
	log.Println("data starting sequence:", dp.Value[8:20])
	log.Println("timestamp", timestamp, "MS:", (timestamp / 1000000), dp.Value[:8])

	ParseDataPoint(dp.Value[8:])
	timeTag := strconv.FormatInt(dp.Timestamp, 10)
	// for _, v := range dp.Value {
	if LiveBuffer.Map[dp.Tag] == nil {
		LiveBuffer.Map[dp.Tag] = make(map[string]map[string][][]byte)
	}
	if LiveBuffer.Map[dp.Tag]["meow"] == nil {
		LiveBuffer.Map[dp.Tag]["meow"] = make(map[string][][]byte)
	}

	LiveBuffer.Mux.Lock()
	LiveBuffer.Map[dp.Tag]["meow"][timeTag] = append(LiveBuffer.Map[dp.Tag]["meow"][timeTag], dp.Value)
	LiveBuffer.Mux.Unlock()
	// }
	// log.Println(LiveBuffer.Map)
	for _, v := range LiveBuffer.Map {
		for _, iv := range v {
			for iii, iiv := range iv {
				log.Println(iii, iiv)
			}
		}
	}
}

// func (c *Controller) WriteBufferToFile() {

// 	now := time.Now().Format(time.RFC3339Nano)
// 	now = strings.Replace(now, "-", "/", -1)
// 	now = strings.Replace(now, "T", "/", -1)
// 	now = strings.Replace(now, ":", "/", 1)
// 	//	helpers.DebugLog(strings.Split(now, ":")[0])
// 	err := os.MkdirAll(c.BufferDirectoryPath+strings.Split(now, ":")[0], 0700)
// 	now = strings.Replace(now, ":", "/", 1)
// 	helpers.PanicX(err)
// 	helpers.DebugLog("writing to file:", c.BufferDirectoryPath+now)
// 	file, err := os.OpenFile(c.BufferDirectoryPath+now, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
// 	helpers.PanicX(err)
// 	for i, v := range c.CollectionBuffer {
// 		// file.Write(v. )
// 		for ii, iv := range v {
// 			log.Println("FileName:", iv.Tag)
// 			// 1. open file by tag/datapoint/timestamp
// 			// 2. write contents

// 		}
// 	}
// 	// buffer.WriteTo(file)
// }
