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
func (d *NetworkDynamic) GetFormattedBytes(basePoint bool) []byte {

	var mainList []byte

	base := History.DynamicBasePoint.NetworkDynamic
	mainList = append(mainList, byte(len(d.Interfaces)))
	// changeCount := 0
	for i, v := range d.Interfaces {
		var valueList []int64
		// valueList = append(valueList, []byte(v.Name)...)
		if basePoint {
			valueList = append(valueList, int64(v.IN.Bytes))
			valueList = append(valueList, int64(v.IN.Packets))
			valueList = append(valueList, int64(v.IN.Errors))
			valueList = append(valueList, int64(v.IN.Dropped))
			valueList = append(valueList, int64(v.IN.Frame))
			valueList = append(valueList, int64(v.IN.Compressed))
			valueList = append(valueList, int64(v.IN.Multicast))
			valueList = append(valueList, int64(v.OUT.Bytes))
			valueList = append(valueList, int64(v.OUT.Packets))
			valueList = append(valueList, int64(v.OUT.Errors))
			valueList = append(valueList, int64(v.OUT.Dropped))
			valueList = append(valueList, int64(v.OUT.Frame))
			valueList = append(valueList, int64(v.OUT.Compressed))
			valueList = append(valueList, int64(v.OUT.Multicast))
		} else {
			valueList = append(valueList, int64(v.IN.Bytes)-int64(base.Interfaces[i].IN.Bytes))
			valueList = append(valueList, int64(v.IN.Packets)-int64(base.Interfaces[i].IN.Packets))
			valueList = append(valueList, int64(v.IN.Errors)-int64(base.Interfaces[i].IN.Errors))
			valueList = append(valueList, int64(v.IN.Dropped)-int64(base.Interfaces[i].IN.Dropped))
			valueList = append(valueList, int64(v.IN.Frame)-int64(base.Interfaces[i].IN.Frame))
			valueList = append(valueList, int64(v.IN.Compressed)-int64(base.Interfaces[i].IN.Compressed))
			valueList = append(valueList, int64(v.IN.Multicast)-int64(base.Interfaces[i].IN.Multicast))
			valueList = append(valueList, int64(v.OUT.Bytes)-int64(base.Interfaces[i].OUT.Bytes))
			valueList = append(valueList, int64(v.OUT.Packets)-int64(base.Interfaces[i].OUT.Packets))
			valueList = append(valueList, int64(v.OUT.Errors)-int64(base.Interfaces[i].OUT.Errors))
			valueList = append(valueList, int64(v.OUT.Dropped)-int64(base.Interfaces[i].OUT.Dropped))
			valueList = append(valueList, int64(v.OUT.Frame)-int64(base.Interfaces[i].OUT.Frame))
			valueList = append(valueList, int64(v.OUT.Compressed)-int64(base.Interfaces[i].OUT.Compressed))
			valueList = append(valueList, int64(v.OUT.Multicast)-int64(base.Interfaces[i].OUT.Multicast))
		}
		ifData := helpers.WriteValueList(valueList, v.Name)
		mainList = append(mainList, ifData...)
	}
	log.Println("FINAL NETWORK RETURN")
	log.Println(mainList)
	return mainList
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
