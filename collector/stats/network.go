package stats

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"strings"

	gonet "github.com/shirou/gopsutil/net"
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
	Name      string
	IN        *IN
	OUT       *OUT
	ValueList []int64
}

type NetworkDynamic struct {
	Interfaces map[string]*NetworkInterface
	ValueList  []int64
}

type NetworkStatic struct {
	Name               string
	HardwareAddress    string
	Addresses          []string
	MulticastAddresses []string
	Flags              string
	MTU                int
	Index              int
}

func GetNetworkBytes(history *HistoryBuffer) []byte {
	var data []byte
	netstuff, err := gonet.IOCounters(true)
	helpers.PanicX(err)
	for _, v := range netstuff {
		History.MinimumStats.NetworkIn = History.MinimumStats.NetworkIn + v.BytesRecv
		History.MinimumStats.NetworkOut = history.MinimumStats.NetworkOut + v.BytesSent
	}

	// return nothing if we have no previous history
	if History.PreviousMinimumStats == nil {
		// log.Println("SAING NETWORK:", 0, 0, data)
		return []byte{byte(0), byte(0)}
	}

	// bytes per second IN
	InPerSecond := history.MinimumStats.NetworkIn - history.PreviousMinimumStats.NetworkIn
	OutPerSecond := history.MinimumStats.NetworkOut - history.PreviousMinimumStats.NetworkOut
	// log.Println("N I/O", InPerSecond, OutPerSecond)
	var inBuffer = new(bytes.Buffer)
	lengthIN := helpers.WriteIntToBuffer(inBuffer, int64(InPerSecond))
	data = append(data, byte(lengthIN))
	data = append(data, inBuffer.Bytes()...)
	var outBuffer = new(bytes.Buffer)
	LengthOut := helpers.WriteIntToBuffer(outBuffer, int64(OutPerSecond))
	data = append(data, byte(LengthOut))
	data = append(data, outBuffer.Bytes()...)
	// log.Println("POST NETWORK:", data)
	// binary.Write(buffer, binary.LittleEndian, int32(InPerSecond))
	// binary.Write(buffer, binary.LittleEndian, int32(OutPerSecond))

	// log.Println("SAING NETWORK:", InPerSecond, OutPerSecond, data)
	return data
}

func collectNetworkStats(sp *StaticPoint) {
	netIF, err := gonet.Interfaces()
	NetworkStaticList := make(map[string]*NetworkStatic)
	helpers.PanicX(err)
	for _, v := range netIF {

		var networkAddressLis []string

		for _, av := range v.Addrs {
			networkAddressLis = append(networkAddressLis, av.String())
		}

		NetworkStaticList[v.Name] = &NetworkStatic{
			Name:            v.Name,
			HardwareAddress: v.HardwareAddr,
			Addresses:       networkAddressLis,
			Flags:           strings.Join(v.Flags, ","),
			MTU:             v.MTU,
		}
	}
	sp.NetworkStatic = NetworkStaticList
}
func collectNetworkData() map[string]*NetworkInterface {

	netstuff, err := gonet.IOCounters(true)

	helpers.PanicX(err)
	NIFL := make(map[string]*NetworkInterface)
	var NIF *NetworkInterface
	for _, v := range netstuff {

		NIF = &NetworkInterface{
			Name: v.Name,
			IN:   &IN{},
			OUT:  &OUT{},
		}
		// in
		NIF.IN.Bytes = float64(v.BytesRecv)
		NIF.IN.Dropped = float64(v.Dropin)
		NIF.IN.Errors = float64(v.Errin)
		NIF.IN.Fifo = float64(v.Fifoin)
		NIF.IN.Packets = float64(v.PacketsRecv)
		// out
		NIF.OUT.Bytes = float64(v.BytesSent)
		NIF.OUT.Dropped = float64(v.Dropout)
		NIF.OUT.Errors = float64(v.Errout)
		NIF.OUT.Fifo = float64(v.Fifoout)
		NIF.OUT.Packets = float64(v.PacketsSent)
		NIFL[v.Name] = NIF
	}
	return NIFL
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
			// Flags:              uint(nif.Flags),
			MTU:   nif.MTU,
			Index: nif.Index,
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
		// composition = append(composition, ""+strings.Join(nif.MulticastAddresses, ","))
		composition = append(composition, ""+nif.HardwareAddress)
		composition = append(composition, ""+nif.Flags)
		// composition = append(composition, ""+strconv.Itoa(nif.Index))
		composition = append(composition, ""+strconv.Itoa(nif.MTU))
	}
	return strings.Join(composition, ",")
}

func collectNetworkDownloadAndUpload(dp *DynamicPoint) {
	// networkInterfaces := getAllInterfacesDownloadAndUploadStats()
	networkInterfaces := collectNetworkData()
	dp.NetworkDynamic = &NetworkDynamic{Interfaces: networkInterfaces}
}
func (d *NetworkDynamic) GetFormattedBytes(basePoint bool) []byte {

	var mainList []byte

	// base := History.DynamicBasePoint.NetworkDynamic
	// prev := History.PreviousDynamicUpdatePoint

	mainList = append(mainList, byte(len(d.Interfaces)))
	// changeCount := 0
	for _, v := range d.Interfaces {
		var valueList []int64
		// valueList = append(valueList, []byte(v.Name)...)
		if basePoint {
			base := History.DynamicBasePoint.NetworkDynamic.Interfaces[v.Name]
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
			ifData := helpers.WriteValueList(valueList, v.Name)
			mainList = append(mainList, ifData...)
			base.ValueList = valueList
		} else {
			prev := History.DynamicPreviousUpdatePoint.NetworkDynamic.Interfaces[v.Name]
			base := History.DynamicBasePoint.NetworkDynamic.Interfaces[v.Name]
			v.ValueList = append(v.ValueList, int64(v.IN.Bytes))
			v.ValueList = append(v.ValueList, int64(v.IN.Packets))
			v.ValueList = append(v.ValueList, int64(v.IN.Errors))
			v.ValueList = append(v.ValueList, int64(v.IN.Dropped))
			v.ValueList = append(v.ValueList, int64(v.IN.Frame))
			v.ValueList = append(v.ValueList, int64(v.IN.Compressed))
			v.ValueList = append(v.ValueList, int64(v.IN.Multicast))
			v.ValueList = append(v.ValueList, int64(v.OUT.Bytes))
			v.ValueList = append(v.ValueList, int64(v.OUT.Packets))
			v.ValueList = append(v.ValueList, int64(v.OUT.Errors))
			v.ValueList = append(v.ValueList, int64(v.OUT.Dropped))
			v.ValueList = append(v.ValueList, int64(v.OUT.Frame))
			v.ValueList = append(v.ValueList, int64(v.OUT.Compressed))
			v.ValueList = append(v.ValueList, int64(v.OUT.Multicast))
			ifData := helpers.WriteValueList2(v.ValueList, base.ValueList, prev.ValueList, v.Name)
			mainList = append(mainList, ifData...)

		}
	}
	// log.Println("FINAL NETWORK RETURN")
	// log.Println(mainList)
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
