package alerting

import "github.com/zkynetio/safelocker"

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
	DefaultType string  `json:"default_type"`
	Default     []Alert `json:"default"`
	Custom      []Alert `json:"custom"`
}

type Alert struct {
	Tag       string   `json:"tag"`
	Namespace string   `json:"namespace"`
	Value     int      `json:"value"`
	Time      string   `json:"time"`
	Count     int      `json:"count"`
	Color     string   `json:"color"`
	To        []string `json:"to"`
}

type ActiveAlert struct {
	Namespace string
	Count     int
	Triggerd  uint64
	Alert     *Alert
}
type AlertBucket struct {
	safelocker.SafeLocker
	// tag // namespace // count // first trigger ( time )
	ActiveAlert map[string]ActiveAlert
}

func (a *AlertBucket) AddAlert(alert ActiveAlert, tag string) {
	a.Lock()
	defer a.Unlock()
	a.ActiveAlert[tag] = alert
}
func (a *AlertBucket) RemoveAlert(tag string) {
	a.Lock()
	defer a.Unlock()
	delete(a.ActiveAlert, tag)
}
