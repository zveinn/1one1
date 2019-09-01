package brain

import (
	"net"

	"github.com/zkynetio/lynx/alerting"
)

type Brain struct {
	Config      Config                    `json:"-"`
	Alerting    []alerting.Alerting       `json:"alerting"`
	Collecting  Collecting                `json:"collecting"`
	Controllers map[string]LiveController `json:"-"`
}
type LiveController struct {
	Socket net.Conn    `json:"-"`
	Config *Controller `json:"config"`
}

type Collecting struct {
	Default []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"default"`
	Custom []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"custom"`
}

type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
	UI   struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"ui"`
	User            string `json:"user"`
	Pass            string `json:"pass"`
	Clusters        []Cluster
	AlertingConfigs []string
}
type Cluster struct {
	Tag         string `json:"tag"`
	Controllers []Controller
}
type Controller struct {
	IP        string          `json:"ip"`
	UI        UIConfig        `json:"ui"`
	Collector CollectorConfig `json:"collector"`
	Live      bool
	Shutdown  bool `json:"shutdown"`
	Debug     bool `json:"debug"`
	Restart   bool `json:"restart"`
}
type CollectorConfig struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
type UIConfig struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
