package stats

import (
	"bytes"
	"strconv"
	"strings"

	gonet "github.com/shirou/gopsutil/net"
	"github.com/zkynetio/lynx/helpers"
)

type NetworkInterface struct {
	Name      string
	ValueList []int64
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

	if History.PreviousMinimumStats == nil {
		return []byte{byte(0), byte(0)}
	}

	InPerSecond := history.MinimumStats.NetworkIn - history.PreviousMinimumStats.NetworkIn
	OutPerSecond := history.MinimumStats.NetworkOut - history.PreviousMinimumStats.NetworkOut
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
