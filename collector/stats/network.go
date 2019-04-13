package stats

import (
	"io/ioutil"
	"log"
	"net"
	"strings"

	"github.com/zkynetio/lynx/helpers"
)

type NetworkDynamic struct {
	DownlodBytes    float64
	DownloadPackets float64
	UploadBytes     float64
	UploadPackets   float64
}
type NetworkStatic struct {
}
type NetworkInterface struct {
	Name               string
	HardwareAddress    []byte
	Addresses          []net.Addr
	MulticastAddresses []net.Addr
	Flags              uint
	MTU                int
	Index              int
}

func FetchNetworkIFS() {
	interfStat, err := net.Interfaces()
	helpers.PanicX(err)

	for _, interf := range interfStat {

		log.Println("Name:", interf.Name)
		log.Println("HardwareAddr:", interf.HardwareAddr)
		addrs, err := interf.Addrs()
		helpers.PanicX(err)
		for _, addr := range addrs {
			log.Println("ADDR:::", addr)
		}
		maddrs, err := interf.MulticastAddrs()
		helpers.PanicX(err)
		for _, addr := range maddrs {
			log.Println("MULTICAST ADDS:::", addr)
		}
		log.Println("flags", interf.Flags)
		log.Println("mtu", interf.MTU)
		log.Println("index", interf.Index)
		//for _, flag := range interf.Flags {
		//	log.Println("FLAG:::", flag)
		//}
		//getUploadDownload(interf.Name)

	}
}

func getUploadDownload(ifname string) {
	file, err := ioutil.ReadFile("/proc/net/dev") // O_RDONLY mode
	if err != nil {
		log.Fatal(err)
	}
	//defer file.Close()
	filestring := string(file)
	//log.Println("=============================")
	//log.Println(filestring)
	//log.Println("=============================")
	fileSplit := strings.Split(filestring, "\n")
	for _, v := range fileSplit {
		if strings.Contains(v, ifname) {
			vSplit := strings.Split(v, " ")
			//for i, v := range vSplit {
			//	log.Println(i, ":", v)
			//}
			log.Println("Download bytes:", vSplit[1])
			log.Println("Download packets:", vSplit[2])
			log.Println("Upload bytes:", vSplit[39])
			log.Println("Upload packates:", vSplit[41])
		}
		//if v == ifname {
		//	log.Println(i, ":", v)
		//	return "FOUND IT!"
		//}
	}
	//log.Println(fileSplit[247])

}
