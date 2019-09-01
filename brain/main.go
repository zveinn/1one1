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

	"github.com/google/go-cmp/cmp"
	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/lynx/helpers"
)

var GlobalBrain *Brain

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
	for {
		time.Sleep(5 * time.Second)
		ReadBrainConfigAtStart(b)
	}
}
func ReadBrainConfigAtStart(b *Brain) {
	file, err := ioutil.ReadFile("brain.json")
	if err != nil {
		panic(err)
	}
	data := Config{}
	_ = json.Unmarshal([]byte(file), &data)
	configSent := false
	oldData := b.Config
	b.Config = data
	if !cmp.Equal(b.Config.AlertingConfigs, oldData.AlertingConfigs) {
		log.Println("config has changed !!")
		for _, v := range b.Controllers {
			b.SendBrainToController(v)
		}
		configSent = true
	}

	if !configSent {
		// find a way to check controllers for change !
	}

	if data.IP != data.IP || data.Port != data.Port {
		b.SendNewAddressToControllers(data.IP + ":" + strconv.Itoa(data.Port))
	}

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
	ReadBrainConfigAtStart(&Brain)
	go ReadBrainConfig(&Brain)
	ReadAlertingConfig(&Brain)
	ReadCollectionConfig(&Brain)
	GlobalBrain = &Brain

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
func (b *Brain) AddController(LC *LiveController) {
	b.Lock()
	b.Controllers[LC.Socket.RemoteAddr().String()] = LC
	defer b.Unlock()
}
func (b *Brain) RemoveController(address string) {
	b.Lock()
	defer b.Unlock()
	delete(b.Controllers, address)
}
func (b *Brain) ListenForControllers(watcher chan int) {
	b.Controllers = make(map[string]*LiveController)
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

		go b.acceptController(socket)
	}
}
func (b *Brain) acceptController(socket net.Conn) {
	LC := LiveController{
		Socket: socket,
	}
	err := b.AssignControllerToIPAndPort(&LC)
	if err != nil {
		log.Println("Could not assign controller to an IP and PORT", err)
		socket.Close()
		return
	}
	log.Println("found config", LC.Config)
	b.SendConfigToController(&LC)
	b.SendBrainToController(&LC)
	b.AddController(&LC)
	LC.ListenToController()
}
func (b *Brain) SendConfigToController(LC *LiveController) {

	jsonConfig, err := json.Marshal(LC.Config)
	if err != nil {
		panic(err)
	}

	// go func() {
	// 	for {
	// 		time.Sleep(20000 * time.Millisecond)
	// 		log.Println("sending config again!")
	// 		var data = new(bytes.Buffer)
	// 		binary.Write(data, binary.LittleEndian, uint16(len(jsonConfig)))
	// 		data.Write([]byte{0x01})
	// 		data.Write(jsonConfig)
	// 		_, err := data.WriteTo(LC.Socket)
	// 		if err != nil {
	// 			return
	// 		}
	// 	}
	// }()

	log.Println("Sending config;", string(jsonConfig))
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(jsonConfig)))
	data.Write([]byte{0x01})
	data.Write(jsonConfig)
	// log.Println("config bytes:", data.Bytes())
	data.WriteTo(LC.Socket)
}
func (b *Brain) SendBrainToController(LC *LiveController) {
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
func (b *Brain) SendNewAddressToControllers(address string) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, uint16(len(address)))
	data.Write([]byte{0x03})
	log.Println("Sending a new address to controllers.")
	data.Write([]byte(address))
	outData := data.Bytes()
	for _, v := range b.Controllers {
		_, err := v.Socket.Write(outData)
		if err != nil {
			log.Println("Could not send new address to controller:", v.Config.IP, v.Config.Collector.Port)
			continue
		}
	}
}
func (c *LiveController) ListenToController() {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("Panic in controller reader", r, string(debug.Stack()))
		}
		GlobalBrain.UnAssignControllerFromIPAndPort(c)
		GlobalBrain.RemoveController(c.Socket.RemoteAddr().String())
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
func (b *Brain) UnAssignControllerFromIPAndPort(c *LiveController) {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("Panic when un assigning controller", r, string(debug.Stack()))
			b.SafeUnlock()
		}
	}()
	b.Lock()
	for i, v := range b.Config.Clusters {
		for ii, iv := range v.Controllers {
			if iv.IP == c.Config.IP && iv.Collector.Port == c.Config.Collector.Port {
				log.Println("UnAssign Controller:", iv.IP, iv.Live, iv.Collector.Port)
				b.Config.Clusters[i].Controllers[ii].Live = false
				break
			}
		}
	}
	b.Unlock()

}
func (b *Brain) AssignControllerToIPAndPort(c *LiveController) error {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("Panic when un assigning controller", r, string(debug.Stack()))
			b.SafeUnlock()
		}
	}()
	b.Lock()
	cIP := strings.Split(c.Socket.RemoteAddr().String(), ":")[0]
	err := errors.New("could not find a config for IP:" + cIP)
	for i, v := range b.Config.Clusters {
		for ii, iv := range v.Controllers {
			if iv.IP == cIP {
				log.Println("Assigning Controller:", iv.IP, iv.Live, iv.Collector.Port)
				err = errors.New("all controlles assigned to that IP address are already live")
				if !iv.Live {
					err = nil
					b.Config.Clusters[i].Controllers[ii].Live = true
					c.Config = &iv
					break
				}
			}
		}
	}
	b.Unlock()
	return err
}
