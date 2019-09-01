package brain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
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
func Start() {

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
			time.Sleep(1 * time.Second)
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
			log.Println("Receiver crashed", r, string(debug.Stack()))
		}
		watcher <- 1
	}(watcher)

	log.Println("listening on:", b.Config.IP+":"+strconv.Itoa(b.Config.Port))
	ln, err := net.Listen("tcp", b.Config.IP+":"+strconv.Itoa(b.Config.Port))
	helpers.PanicX(err)
	for {
		socket, err := ln.Accept()
		helpers.PanicX(err)
		LC := LiveController{
			Socket: socket,
		}
		err = b.AssignControllerToIPAndPort(&LC)
		log.Println("Assignment error:", err)
		if err != nil {
			errLength := len(err.Error())
			data := new(bytes.Buffer)
			binary.Write(data, binary.LittleEndian, uint16(errLength))
			data.Write([]byte{0x00})
			data.Write([]byte(err.Error()))
			data.WriteTo(socket)
			socket.Close()
			continue
		}

		go b.acceptController(LC)
	}
}
func (b *Brain) acceptController(LC LiveController) {
	log.Println("found config", LC.Config)
	LC.SendConfigToController()
	LC.SendBrainToController(b)
	LC.ListenToController()
}
func (LC *LiveController) SendConfigToController() {

	// go func() {
	// 	for {
	// 		time.Sleep(20000 * time.Millisecond)
	// 		log.Println("sending config again!")
	// 		var data = new(bytes.Buffer)
	// 		binary.Write(data, binary.LittleEndian, uint16(len(con)))
	// 		data.Write(con)
	// 		_, err := data.WriteTo(c.Socket)
	// 		if err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	jsonConfig, err := json.Marshal(LC.Config)
	if err != nil {
		panic(err)
	}

	log.Println("Sending config;", string(jsonConfig))
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(jsonConfig)))
	data.Write([]byte{0x01})
	data.Write(jsonConfig)
	// log.Println("config bytes:", data.Bytes())
	data.WriteTo(LC.Socket)
}
func (LC *LiveController) SendBrainToController(b *Brain) {
	jsonBrain, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	log.Println("Sending brain;", string(jsonBrain))
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(jsonBrain)))
	data.Write([]byte{0x02})
	data.Write(jsonBrain)
	data.WriteTo(LC.Socket)

}

func (c *LiveController) ListenToController() {

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

func (b *Brain) AssignControllerToIPAndPort(c *LiveController) error {
	cIP := strings.Split(c.Socket.RemoteAddr().String(), ":")[0]
	err := errors.New("could not find a config for IP:" + cIP)
	for i, v := range b.Config.Clusters {
		for ii, iv := range v.Controllers {
			log.Println("Controller:", iv.IP, iv.Live, iv.Collector.Port)
			if iv.IP == cIP {
				err = errors.New("all controlles assigned to that IP address are already live")
				if !iv.Live {
					err = nil
					b.Lock()
					b.Config.Clusters[i].Controllers[ii].Live = true
					b.Unlock()
					c.Config = &iv
					break
				}
			}
		}
	}

	return err
}
