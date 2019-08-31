package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/lynx/helpers"
)

func ReadCollectionConfig(b *Brain) {
	file, err := ioutil.ReadFile("collecting.json")
	if err != nil {
		panic(err)
	}
	data := Collecting{}
	_ = json.Unmarshal([]byte(file), &data)
	b.Collecting = data
}
func ReadBrainConfig(b *Brain) {
	file, err := ioutil.ReadFile("brain.json")
	if err != nil {
		panic(err)
	}
	data := Config{}
	_ = json.Unmarshal([]byte(file), &data)
	b.Config = data
}
func ReadAlertingConfig(b *Brain) {

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if strings.Contains(f.Name(), "alerts") {
			file, err := ioutil.ReadFile(f.Name())
			if err != nil {
				panic(err)
			}

			data := alerting.Alerting{}
			_ = json.Unmarshal([]byte(file), &data)
			b.Alerting = append(b.Alerting, data)
			// b.Alerting[data.Name] = data
		}
	}
}
func main() {

	Brain := Brain{}
	ReadBrainConfig(&Brain)
	ReadAlertingConfig(&Brain)
	ReadCollectionConfig(&Brain)

	log.Println(Brain.Config)
	log.Println(Brain.Alerting)
	log.Println(Brain.Collecting)

	watcherChannel := make(chan int)
	stop := make(chan os.Signal, 1)
	go Brain.ListenForControllers(watcherChannel)

	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case index := <-watcherChannel:
			if index == 1 {
				go Brain.ListenForControllers(watcherChannel)
			}
			log.Println("goroutine number", index, "just closed...")
			break
		case <-stop:
			// TODO: handle exit gracefully
			log.Println("handle exit gracefully")
			os.Exit(1)
		}
	}

}
func (b *Brain) ListenForControllers(watcher chan int) {
	b.Controllers = make(map[string]LiveController)
	defer func(watcher chan int) {
		if r := recover(); r != nil {
			log.Println("Receiver crashed", r)
		}
		watcher <- 1
	}(watcher)

	helpers.DebugLog("listening on:", b.Config.IP+":"+strconv.Itoa(b.Config.Port))
	ln, err := net.Listen("tcp", b.Config.IP+":"+strconv.Itoa(b.Config.Port))
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		helpers.PanicX(err)
		go b.acceptController(conn)
	}
}
func (b *Brain) acceptController(socket net.Conn) {
	LC := LiveController{
		Socket: socket,
	}
	b.AssignControllerToIPAndPort(&LC)
	log.Println("found config", LC.Config)

	jsonConfig, err := json.Marshal(LC.Config)
	if err != nil {
		panic(err)
	}

	log.Println("Sending config;", string(jsonConfig))
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(jsonConfig)))
	data.Write(jsonConfig)
	data.WriteTo(socket)

	jsonBrain, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	log.Println("Sending brain;", string(jsonBrain))
	data = new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(jsonBrain)))
	data.Write(jsonBrain)
	data.WriteTo(socket)

	LC.ListenToController(jsonConfig)
}
func (c *LiveController) ListenToController(con []byte) {

	go func() {
		for {
			time.Sleep(20000 * time.Millisecond)
			log.Println("sending config again!")
			var data = new(bytes.Buffer)
			binary.Write(data, binary.LittleEndian, uint16(len(con)))
			data.Write(con)
			data.WriteTo(c.Socket)
		}
	}()
	buf := make([]byte, 20000)
	for {
		n, err := c.Socket.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("FROM CONTROLLER:", buf[0:n])
	}
}

func (b *Brain) AssignControllerToIPAndPort(c *LiveController) {
	cIP := strings.Split(c.Socket.RemoteAddr().String(), ":")[0]
	for _, v := range b.Config.Clusters {
		for _, iv := range v.Controllers {
			if iv.IP == cIP {
				iv.Live = true
				c.Config = &iv
			}
		}
	}
}
