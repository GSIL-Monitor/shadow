package gate

import (
	"strconv"
	"strings"
	// third pkgs
	log "github.com/cihub/seelog"
	yaml "gopkg.in/yaml.v2"
	// company pkgs
)

// ------------------------------------------------------
// Database model -- Begin
type Ring struct {
	Id   uint16
	Name string
}
type App struct {
	Id   uint16
	Name string
}
type Instance struct {
	Id     uint32
	Ip     uint64
	Port   uint32
	Weight uint32
	RingId uint16 // logical Foreign Key
}

// Database model -- End
// ------------------------------------------------------

// ------------------------------------------------------
// YAML -- Begin
type AppNames []string
type Migration struct {
} //TODO extension
type Shard struct {
	// IP:Port,Weight
	Master string   `yaml:"master"`
	Slaves []string `yaml:"slaves,omitempty"`
}
type RingYAML struct {
	Name            string     `yaml:"name"`
	Datetime        string     `yaml:"datetime"`
	ProtocolVersion string     `yaml:"protocol_version"`
	Owner           string     `yaml:"owner,omitempty"`
	AlertReceivers  []string   `yaml:"alert_receivers,omitempty"`
	Migration       *Migration `yaml:"migration"`
	AppNames        AppNames   `yaml:"app_names"` // simple ACL implementation
	Shards          []*Shard   `yaml:"shards"`
}

// YAML -- End
// ------------------------------------------------------

func DecodeRingYAMLByStr(ringName, val string) (ring *RingYAML) {
	buf := []byte(val)
	return DecodeRingYAML(ringName, buf)
}
func DecodeRingYAML(ringName string, buf []byte) (ring *RingYAML) {
	ring = &RingYAML{}
	err := yaml.Unmarshal(buf, &ring)
	if err != nil {
		log.Error("Unmarshalling has one error for getting a RingYaml. Error:", err)
		return nil
	}
	ring.Name = ringName
	return
}

func DecodeAppsYAML(cnf []byte) (apps []string) {
	apps = make([]string, 0)
	err := yaml.Unmarshal(cnf, &apps)
	if err != nil {
		log.Error("Unmarshalling has one error for getting a list of app name. Error:", err)
		return nil
	}
	return
}

// Just parse master
func (ring *RingYAML) getInstances() (ins map[string]uint32) {
	ins = make(map[string]uint32)
	for _, shard := range ring.Shards {
		if addr, w := parseStrIns(shard.Master); addr != "" {
			ins[addr] = w
		}
	}
	return
}

// parameter:   IP:Port,Weight
func parseStrIns(val string) (addr string, weight uint32) {
	val = strings.TrimSpace(val)
	a := strings.Split(val, ",")
	addr = strings.TrimSpace(a[0])
	w, err := strconv.Atoi(strings.TrimSpace(a[1]))
	if err != nil {
		log.Error("Parse Error:", val)
		return
	}
	if w <= 0 {
		log.Info("Weight <= 0! The input is: ", val, ". Please check the configuration.")
	}
	weight = uint32(w)
	return
}
