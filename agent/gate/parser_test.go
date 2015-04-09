package gate

import (
	"log"
	"reflect"
	"strconv"
	"testing"
)

const (
	DATETIME_FORMAT = "2006-01-02 15:04"
)

var (
	ringName       = "test"
	test_ring_yaml = `
                    datetime:   2015-01-01 13:00        # be precise to minute
                    protocol_version:   1.0 
                    owner:    zhaoxi
                    alert_receivers:
                    migration:
                    app_names:    
                        - CPC_Ad
                        - Search-Ad
                    shards:
                        - # one shard is one object
                          master: 10.11.1.1:6379,1      # ip:port,weight
                          slaves:
                                - 10.11.2.1:6379,1
                                - 10.11.3.1:6379,1
                        -
                          master: 10.11.1.2:6379,1      # ip:port,weight
                          slaves: []                    # empty content
                        -
                          master: 10.11.1.3:6379,1
                        -
                          master: 10.11.1.10:6379,2
    `
)

// func DecodeRingYAML(ringName string, buf []byte) (ring *RingYAML)
func TestDecodeRingYAML(t *testing.T) {
	expectedRing := newExpectedRing()
	actual := DecodeRingYAML(ringName, []byte(test_ring_yaml))
	assertRing(actual, expectedRing, t)
}

// func DecodeRingYAMLByStr(ringName, val string) (ring *RingYAML)
func TestDecodeRingYAMLByStr(t *testing.T) {
	expectedRing := newExpectedRing()
	actual := DecodeRingYAMLByStr(ringName, test_ring_yaml)
	assertRing(actual, expectedRing, t)
}

// func DecodeAppsYAML(cnf []byte) (apps []string)
func TestDecodeAppsYAML(t *testing.T) {
	data := `
        # A list of app name
        - CPC_1_Shadow
        - CPC_2_Shadow
        - TOP_1_Shadow 
    `
	apps := DecodeAppsYAML([]byte(data))
	if len(apps) != 3 {
		t.Error("Not expected apps length.")
	}
	if apps[0] != "CPC_1_Shadow" {
		t.Error("Not expected AppName. ", apps[0])
	}
	if apps[1] != "CPC_2_Shadow" {
		t.Error("Not expected AppName. ", apps[1])
	}
	if apps[2] != "TOP_1_Shadow" {
		t.Error("Not expected AppName. ", apps[2])
	}
}

// func (ring *RingYAML) getInstances() (ins map[string]uint32)
func TestGetInstances(t *testing.T) {
	r := DecodeRingYAMLByStr(ringName, test_ring_yaml)
	expected := make(map[string]uint32)
	expected["10.11.1.1:6379"] = 1
	expected["10.11.1.2:6379"] = 1
	expected["10.11.1.3:6379"] = 1
	expected["10.11.1.10:6379"] = 2
	assertMap(expected, r.getInstances(), t)
}

// func parseStrIns(val string) (addr string, weight uint32)
func TestParseStrIns(t *testing.T) {
	addr := "127.0.0.1:6379"
	weight := "10"
	val := addr + "," + weight
	a, w := parseStrIns(val)
	i, err := strconv.Atoi(weight)
	if err != nil {
		t.Error("Error string to int", weight)
	}
	if a != addr {
		t.Error("parse error")
	}
	if w != uint32(i) {
		t.Error("parse error")
	}
}

func newExpectedRing() (expectedRing *RingYAML) {
	shards := make([]*Shard, 0)
	shard := &Shard{
		Master: "10.11.1.1:6379,1",
		Slaves: []string{"10.11.2.1:6379,1", "10.11.3.1:6379,1"},
	}
	shards = append(shards, shard)
	shard = &Shard{
		Master: "10.11.1.2:6379,1",
		Slaves: make([]string, 0),
	}
	shards = append(shards, shard)
	shard = &Shard{
		Master: "10.11.1.3:6379,1",
		Slaves: nil,
	}
	shards = append(shards, shard)
	shard = &Shard{
		Master: "10.11.1.10:6379,2",
		Slaves: nil,
	}
	shards = append(shards, shard)
	expectedRing = &RingYAML{
		Name:            ringName,
		Datetime:        "2015-01-01 13:00",
		ProtocolVersion: "1.0",
		Owner:           "zhaoxi",
		AlertReceivers:  nil,
		Migration:       nil,
		AppNames:        []string{"CPC_Ad", "Search-Ad"},
		Shards:          shards,
	}
	return
}
func assertRing(actual, expectedRing *RingYAML, t *testing.T) {
	if actual == nil {
		t.Errorf("DecodeRingYAML fail...")
	}
	actual.Name = "test"

	if actual.Shards[1].Slaves == nil {
		log.Println("Nil pointer!")
	}

	if !reflect.DeepEqual(actual, expectedRing) {
		t.Errorf("Not equal. Expected:%q, Actual:%q\n", expectedRing, actual)
	}
}

func assertMap(m1, m2 map[string]uint32, t *testing.T) {
	if len(m1) != len(m2) {
		t.Error("Not equal length.")
	}
	for k1, v1 := range m1 {
		v2, ok := m2[k1]
		if !ok {
			t.Error("Not exist key-value. Key:", k1)
		}
		if v1 != v2 {
			t.Errorf("Not expected value. v1:%q , v2:%q.", v1, v2)
		}
	}
}
