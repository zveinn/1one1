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

type NetworkDynamic struct {
	Interfaces map[string]*NetworkInterface
}

type NetworkStatic struct {
	Name               string
	HardwareAddress    string
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
			HardwareAddress:    nif.HardwareAddr.String(),
			Addresses:          networkAddressLis,
			MulticastAddresses: multiAddressList,
			Flags:              uint(nif.Flags),
			MTU:                nif.MTU,
			Index:              nif.Index,
		}
		NetworkStaticList[nif.Name] = NetworkStaticPoint
	}

	sp.NetworkStatic = NetworkStaticList
}

func getFormattedStringForInterfaces(nsl map[string]*NetworkStatic) string {

	var composition []string
	for name, nif := range nsl {
		composition = append(composition, "N||"+name)
		composition = append(composition, ""+strings.Join(nif.Addresses, ","))
		composition = append(composition, ""+strings.Join(nif.MulticastAddresses, ","))
		composition = append(composition, ""+nif.HardwareAddress)
		composition = append(composition, ""+strconv.Itoa(int(nif.Flags)))
		composition = append(composition, ""+strconv.Itoa(nif.Index))
		composition = append(composition, ""+strconv.Itoa(nif.MTU))
	}
	return strings.Join(composition, ",")
}

func collectNetworkDownloadAndUpload(dp *DynamicPoint) {
	networkInterfaces := getAllInterfacesDownloadAndUploadStats()
	dp.NetworkDynamic = NetworkDynamic{Interfaces: networkInterfaces}
}
func (nd *NetworkDynamic) GetFormattedString() string {
	var interfaceslice []string
	if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces != nil {
		changeCount := 0

		for i, v := range nd.Interfaces {

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Bytes != v.IN.Bytes {
				interfaceslice = append(interfaceslice, "ib:"+strconv.Itoa(int(v.IN.Bytes-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Bytes)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Packets != v.IN.Packets {
				interfaceslice = append(interfaceslice, "ip:"+strconv.Itoa(int(v.IN.Packets-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Packets)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Errors != v.IN.Errors {
				interfaceslice = append(interfaceslice, "ie:"+strconv.Itoa(int(v.IN.Errors-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Errors)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Dropped != v.IN.Dropped {
				interfaceslice = append(interfaceslice, "id:"+strconv.Itoa(int(v.IN.Dropped-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Dropped)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Frame != v.IN.Frame {
				interfaceslice = append(interfaceslice, "if:"+strconv.Itoa(int(v.IN.Frame-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Frame)))
				changeCount++
			}
			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Compressed != v.IN.Compressed {
				interfaceslice = append(interfaceslice, "ic:"+strconv.Itoa(int(v.IN.Compressed-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Compressed)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Multicast != v.IN.Multicast {
				interfaceslice = append(interfaceslice, "im:"+strconv.Itoa(int(v.IN.Multicast-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].IN.Multicast)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Bytes != v.OUT.Bytes {
				interfaceslice = append(interfaceslice, "ob:"+strconv.Itoa(int(v.OUT.Bytes-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Bytes)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Packets != v.OUT.Packets {
				interfaceslice = append(interfaceslice, "op:"+strconv.Itoa(int(v.OUT.Packets-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Packets)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Errors != v.OUT.Errors {
				interfaceslice = append(interfaceslice, "oe:"+strconv.Itoa(int(v.OUT.Errors-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Errors)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Dropped != v.OUT.Dropped {
				interfaceslice = append(interfaceslice, "od:"+strconv.Itoa(int(v.OUT.Dropped-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Dropped)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Frame != v.OUT.Frame {
				interfaceslice = append(interfaceslice, "of:"+strconv.Itoa(int(v.OUT.Frame-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Frame)))
				changeCount++
			}
			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Compressed != v.OUT.Compressed {
				interfaceslice = append(interfaceslice, "oc:"+strconv.Itoa(int(v.OUT.Compressed-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Compressed)))
				changeCount++
			}

			if History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Multicast != v.OUT.Multicast {
				interfaceslice = append(interfaceslice, "om:"+strconv.Itoa(int(v.OUT.Multicast-History.DynamicPointMap[HighestHistoryIndex-1].NetworkDynamic.Interfaces[i].OUT.Multicast)))
				changeCount++
			}

			if changeCount > 0 {
				interfaceslice = append(interfaceslice, "in:"+v.Name)
			}
			changeCount = 0
		}
	}

	//log.Println(interfaceslice)
	return strings.Join(interfaceslice, ",")

}
func getAllInterfacesDownloadAndUploadStats() map[string]*NetworkInterface {
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
	NIFL := make(map[string]*NetworkInterface)
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
				NIFL[NIF.Name] = NIF

			}

		}

	}

	//log.Println(NIF, NIF.IN, NIF.OUT)
	return NIFL
}
