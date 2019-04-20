package stats

import (
	"strconv"

	"github.com/shirou/gopsutil/host"
	"github.com/zkynetio/lynx/helpers"
)

type HostStatic struct {
	HostID               string
	KernelVersion        string
	OS                   string
	Platform             string
	PlatformFamily       string
	PlatformVersion      string
	Uptime               string
	VirtualizationRole   string
	VirtualizationSystem string
	BootTime             string
}

func collectStaticHostData(sp *StaticPoint) {
	hostStat, err := host.Info()
	helpers.PanicX(err)
	sp.HostStatic = HostStatic{
		HostID:               hostStat.HostID,
		KernelVersion:        hostStat.KernelVersion,
		OS:                   hostStat.OS,
		Platform:             hostStat.Platform,
		PlatformFamily:       hostStat.PlatformFamily,
		PlatformVersion:      hostStat.PlatformVersion,
		Uptime:               strconv.FormatUint(hostStat.Uptime, 10),
		VirtualizationRole:   hostStat.VirtualizationRole,
		VirtualizationSystem: hostStat.VirtualizationSystem,
		BootTime:             strconv.FormatUint(hostStat.BootTime, 10),
	}
}

func GetFormattedString() string {
	hostStat, err := host.Info()
	helpers.PanicX(err)
	host := hostStat.Hostname
	host = host + "," + hostStat.HostID
	host = host + "," + hostStat.KernelVersion
	host = host + "," + hostStat.OS
	host = host + "," + hostStat.Platform
	host = host + "," + hostStat.PlatformFamily
	host = host + "," + hostStat.PlatformVersion
	//	host = host + "," + strconv.FormatUint(hostStat.Procs, 10)
	host = host + "," + strconv.FormatUint(hostStat.Uptime, 10)
	host = host + "," + hostStat.VirtualizationRole
	host = host + "," + hostStat.VirtualizationSystem
	host = host + "," + strconv.FormatUint(hostStat.BootTime, 10)
	return host
}

func GetHost() string {
	hostStat, err := host.Info()
	helpers.PanicX(err)
	host := hostStat.Hostname
	host = host + "," + hostStat.HostID
	host = host + "," + hostStat.KernelVersion
	host = host + "," + hostStat.OS
	host = host + "," + hostStat.Platform
	host = host + "," + hostStat.PlatformFamily
	host = host + "," + hostStat.PlatformVersion
	//	host = host + "," + strconv.FormatUint(hostStat.Procs, 10)
	host = host + "," + strconv.FormatUint(hostStat.Uptime, 10)
	host = host + "," + hostStat.VirtualizationRole
	host = host + "," + hostStat.VirtualizationSystem
	host = host + "," + strconv.FormatUint(hostStat.BootTime, 10)
	return host
}
