package main

import (
	"bufio"
	"bytes"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

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
	Buffer chan string
	// Recovery in sqlite?
	//RecoveryFile string
	BufferDirectoryPath  string
	PORT                 string
	IP                   string
	Collectors           map[string]*Collector
	UIs                  map[string]*UI
	mutex                sync.Mutex
	MaxLinesInBufferFile int
}

func (c *Controller) CleanupOnExit() {
	debugLog("Cleaning up on exit...")
	for _, collector := range c.Collectors {
		if collector.Conn != nil {
			_, _ = collector.Conn.Write([]byte("c\n"))
			debugLog("Closing:", collector.TAG)
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
		debugLog(err)
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

func start(controller *Controller) {

	debugLog("listening on:", controller.IP+":"+controller.PORT)
	ln, err := net.Listen("tcp", controller.IP+":"+controller.PORT)
	panicX(err)
	for {
		conn, err := ln.Accept()
		panicX(err)
		go receiveConnection(conn, controller)
	}
}

func connectUI(ui *UI, controller *Controller, tag string) {
	debugLog("UI:")
	controller.AddUI(tag+"TODORANDOMINT?", ui)
	for {
		outgoing := <-ui.SendChannel
		log.Println(outgoing)
		_, err := ui.Conn.Write([]byte(outgoing))
		//TODO: handle better
		panicX(err)
	}
}

func connectCollector(collector *Collector, controller *Controller, message string) {
	debugLog("COLLECTOR:", string(message))
	collector.TAG = message
	_, err := collector.Conn.Write([]byte("k\n"))
	panicX(err)
	controller.AddCollector(collector.TAG, collector)
	defer func() {
		debugLog("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	log.Println("Starting general collection from:", collector.TAG)
	for {
		message, err := bufio.NewReader(collector.Conn).ReadString('\n')
		if err != nil {
			// TODO: handle better
			debugLog("ERROR IN READ LOOP:", err)
			err = collector.Conn.Close()
			if err != nil {
				// TODO: handle better
				debugLog("ERROR CLOSING INSIDE READ LOOP:", err)
				break
			}
		}
		//TODO: IMPLEMENT SQLITE
		//debugLog("IN:", string(message))
		controller.Buffer <- message
	}
}
func receiveConnection(conn net.Conn, controller *Controller) {
	debugLog("Collector connected:", conn.RemoteAddr())
	message, _ := bufio.NewReader(conn).ReadString('\n')
	if message == "ui\n" {
		connectUI(&UI{Conn: conn}, controller, message)
	} else {
		connectCollector(&Collector{
			LastCheckin: time.Now(),
			Conn:        conn,
		}, controller, message)
	}
}
func (c *Controller) sendToAllUis(msg string) {
	for _, ui := range c.UIs {
		ui.SendChannel <- msg
	}
}

func (c *Controller) EngageBufferFilePipeAndRotation() {
	for {
		time.Sleep(1 * time.Second)
		c.EngageBufferPipe()
	}
}
func (c *Controller) EngageBufferPipe() {

	var buffer bytes.Buffer
	count := 0
	for {
		message := <-c.Buffer
		go c.sendToAllUis(message)
		_, err := buffer.WriteString(message)

		//debugLog("wrote", data)
		if err != nil {
			debugLog(err)
			break
		}

		count++
		debugLog(count)
		if count >= c.MaxLinesInBufferFile {
			now := time.Now().Format(time.RFC3339Nano)
			now = strings.Replace(now, "-", "/", -1)
			now = strings.Replace(now, "T", "/", -1)
			now = strings.Replace(now, ":", "/", 1)
			//	debugLog(strings.Split(now, ":")[0])
			err := os.MkdirAll(c.BufferDirectoryPath+strings.Split(now, ":")[0], 0777)
			now = strings.Replace(now, ":", "/", 1)
			panicX(err)
			file, err := os.OpenFile(c.BufferDirectoryPath+now, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
			panicX(err)
			buffer.WriteTo(file)
			buffer.Reset()
			break
		}
	}
}

func main() {
	rand.Seed(time.Now().Unix())
	loadEnvironmentVariables()

	controller := Controller{
		IP:                   os.Getenv("IP"),
		PORT:                 os.Getenv("PORT"),
		Collectors:           make(map[string]*Collector),
		Buffer:               make(chan string, 10000),
		BufferDirectoryPath:  "./buffers/",
		MaxLinesInBufferFile: 2000,
	}

	defer controller.CleanupOnExit()
	go controller.EngageBufferFilePipeAndRotation()
	start(&controller)

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
