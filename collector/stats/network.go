package stats

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/zkynetio/lynx/helpers"
)

// face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
type IN struct {
	Bytes      float64
	Packets    float64
	Errors     float64
	Dropped    float64
	Fifo       float64
	Frame      float64
	Compressed float64
	Multicast  float64
}
type OUT struct {
	Bytes      float64
	Packets    float64
	Errors     float64
	Dropped    float64
	Fifo       float64
	Frame      float64
	Compressed float64
	Multicast  float64
}
type NetworkInterface struct {
	Name string
	IN   *IN
	OUT  *OUT
}

type NetworkStatic struct {
	Name               string
	HardwareAddress    []byte
	Addresses          []string
	MulticastAddresses []string
	Flags              uint
	MTU                int
	Index              int
}

func collectNetworkInterfaces(sp *StaticPoint) {
	interfStat, err := net.Interfaces()
	helpers.PanicX(err)
	NetworkStaticList := make(map[string]*NetworkStatic)
	for _, nif := range interfStat {
		var networkAddressLis []string
		var multiAddressList []string
		addrs, err := nif.Addrs()
		helpers.PanicX(err)

		for _, v := range addrs {
			networkAddressLis = append(networkAddressLis, v.String())
		}

		maddrs, err := nif.MulticastAddrs()
		helpers.PanicX(err)

		for _, v := range maddrs {
			networkAddressLis = append(networkAddressLis, v.String())
		}
		NetworkStaticPoint := &NetworkStatic{
			Name:               nif.Name,
			HardwareAddress:    nif.HardwareAddr,
			Addresses:          networkAddressLis,
			MulticastAddresses: multiAddressList,
			Flags:              uint(nif.Flags),
			MTU:                nif.MTU,
			Index:              nif.Index,
		}
		NetworkStaticList[nif.Name] = NetworkStaticPoint
	}

	sp.NetworkStaticList = NetworkStaticList
}
func collectNetworkDownloadAndUpload(dp *DynamicPoint) {
	//getUploadDownload("enp0s31f6")
	networkInterfaces := getAllInterfacesDownloadAndUploadStats()
	dp.NetworkDynamic = networkInterfaces
}
func (dn *NetworkInterface) GetFormattedString(index int) {
	var interfaceslice []string

	if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic[index].IN.Bytes != dn.IN.Bytes {
		interfaceslice = append(interfaceslice, strconv.Itoa(int(History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic[index].IN.Bytes-dn.IN.Bytes)))
	} else {
		interfaceslice = append(interfaceslice, "")
	}

}
func getAllInterfacesDownloadAndUploadStats() []*NetworkInterface {
	defer func() {
		r := recover()
		if r != nil {
			log.Println("recovered in network byte stats", r)
		}
	}()
	file, err := ioutil.ReadFile("/proc/net/dev") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	filestring := string(file)
	fileSplit := strings.Split(filestring, "\n")
	var downloadAndUploadList []string
	for i, v := range fileSplit {
		if i == 0 || i == 1 {
			continue
		}
		vSplit := strings.Split(v, " ")
		for _, v := range vSplit {
			if v != "" {
				downloadAndUploadList = append(downloadAndUploadList, v)
			}
		}
	}

	//log.Println(downloadAndUploadList)
	var NIFL []*NetworkInterface
	var NIF *NetworkInterface
	//indexPostName := 0
	for i, v := range downloadAndUploadList {
		if strings.Contains(v, ":") {
			//indexPostName = 0
			NIF = &NetworkInterface{
				Name: strings.Trim(v, ":"),
			}
			NIF.OUT = &OUT{}
			NIF.IN = &IN{}
		} else {
			if i%17 == 1 {
				fv1, _ := strconv.ParseFloat(downloadAndUploadList[i], 32)
				fv2, _ := strconv.ParseFloat(downloadAndUploadList[i+1], 32)
				fv3, _ := strconv.ParseFloat(downloadAndUploadList[i+2], 32)
				fv4, _ := strconv.ParseFloat(downloadAndUploadList[i+3], 32)
				fv5, _ := strconv.ParseFloat(downloadAndUploadList[i+4], 32)
				fv6, _ := strconv.ParseFloat(downloadAndUploadList[i+5], 32)
				fv7, _ := strconv.ParseFloat(downloadAndUploadList[i+6], 32)
				fv8, _ := strconv.ParseFloat(downloadAndUploadList[i+7], 32)
				fv9, _ := strconv.ParseFloat(downloadAndUploadList[i+8], 32)
				fv10, _ := strconv.ParseFloat(downloadAndUploadList[i+9], 32)
				fv11, _ := strconv.ParseFloat(downloadAndUploadList[i+10], 32)
				fv12, _ := strconv.ParseFloat(downloadAndUploadList[i+11], 32)
				fv13, _ := strconv.ParseFloat(downloadAndUploadList[i+12], 32)
				fv14, _ := strconv.ParseFloat(downloadAndUploadList[i+13], 32)
				fv15, _ := strconv.ParseFloat(downloadAndUploadList[i+14], 32)
				fv16, _ := strconv.ParseFloat(downloadAndUploadList[i+15], 32)
				NIF.IN.Bytes = fv1
				NIF.IN.Packets = fv2
				NIF.IN.Errors = fv3
				NIF.IN.Dropped = fv4
				NIF.IN.Fifo = fv5
				NIF.IN.Frame = fv6
				NIF.IN.Compressed = fv7
				NIF.IN.Multicast = fv8

				NIF.OUT.Bytes = fv9
				NIF.OUT.Packets = fv10
				NIF.OUT.Errors = fv11
				NIF.OUT.Dropped = fv12
				NIF.OUT.Fifo = fv13
				NIF.OUT.Frame = fv14
				NIF.OUT.Compressed = fv15
				NIF.OUT.Multicast = fv16
				NIFL = append(NIFL, NIF)

			}

		}

	}

	//log.Println(NIF, NIF.IN, NIF.OUT)
	return NIFL
}
