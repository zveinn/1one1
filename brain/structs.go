package brain

import (
	"net"

	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/safelocker"
)

type Brain struct {
	safelocker.SafeLocker
	Config      Config                     `json:"-"`
	Alerting    []alerting.Alerting        `json:"alerting"`
	Collecting  Collecting                 `json:"collecting"`
	Controllers map[string]*LiveController `json:"-"`
}
type LiveController struct {
	Socket net.Conn    `json:"-"`
	Config *Controller `json:"config"`
}
type CollectionRules struct {
	Tag        string   `json:"tag"`
	Namespaces []string `json:"namespaces"`
}
type Collecting struct {
	Default []CollectionRules `json:"default"`
	Custom  []CollectionRules `json:"custom"`
}

type Config struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
	UI   struct {
		IP   string `json:"ip"`
		Port int    `json:"port"`
	} `json:"ui"`
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
	Live      bool            `json:"-"`
	Shutdown  bool            `json:"shutdown"`
	Debug     bool            `json:"debug"`
	Restart   bool            `json:"restart"`
}
type CollectorConfig struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
type UIConfig struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
