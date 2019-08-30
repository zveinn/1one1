package main

import "net"

type Brain struct {
	Config      Config
	Alerting    []Alerting                `json:"alerting"`
	Collecting  Collecting                `json:"collecting"`
	Controllers map[string]LiveController `json:"-"`
}
type LiveController struct {
	Socket net.Conn    `json:"-"`
	Config *Controller `json:"config"`
}

type Collecting struct {
	Default []struct {
		Tag     string   `json:"tag"`
		Indexes []string `json:"indexes"`
	} `json:"default"`
	Custom []struct {
		Tag     string   `json:"tag"`
		Indexes []string `json:"indexes"`
	} `json:"custom"`
}
type Alerting struct {
	Name  string `json:"name"`
	Slack struct {
	} `json:"slack"`
	Email struct {
	} `json:"email"`
	Irc struct {
	} `json:"irc"`
	Pagerduty struct {
	} `json:"pagerduty"`
	Sms struct {
	} `json:"sms"`
	DefaultType string `json:"default_type"`
	Defaults    []struct {
		Tag       string   `json:"tag"`
		Namespace string   `json:"namespace"`
		Value     int      `json:"value"`
		Time      string   `json:"time"`
		Count     int      `json:"count"`
		Color     string   `json:"color"`
		To        []string `json:"to"`
	} `json:"defaults"`
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
	IP        string    `json:"ip"`
	UI        UI        `json:"ui"`
	Collector Collector `json:"collector"`
	Live      bool
	Shutdown  bool `json:"shutdown"`
	Debug     bool `json:"debug"`
	Restart   bool `json:"restart"`
}
type Collector struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
type UI struct {
	IP   string `json:"ip"`
	Port int    `json:"port"`
}
