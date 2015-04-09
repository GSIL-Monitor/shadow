package cs

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkgs
)

const (
	MD5_CHARS_LEN = 32
)

var (
	INTERVAL  = 10 * time.Second
	listeners = make(map[string]*DefaultListener)
	commander = Commander{running: false}

	keyCompareFunc = func(cnf1, cnf2 interface{}) bool {
		var k1, k2 string
		switch v1 := cnf1.(type) {
		case *SimpleConfiguration:
			{
				k1 = v1.GetKey()
			}
		case *CombinedConfiguration:
			{
				k1 = v1.GetKey()
			}
		default:
			{
				log.Error("Not expected Configuration concrete struct:", cnf1)
				return true
			}
		}
		switch v2 := cnf2.(type) {
		case *SimpleConfiguration:
			{
				k2 = v2.GetKey()
			}
		case *CombinedConfiguration:
			{
				k2 = v2.GetKey()
			}
		default:
			{
				log.Error("Not expected Configuration concrete struct:", cnf2)
				return true
			}
		}
		return k1 < k2
	}
)

const (
	EMPTY_STR_MD5 = "d41d8cd98f00b204e9800998ecf8427e"
)

type Commander struct {
	running bool
	mu      sync.Mutex
}

func init() {
	log.Info("Initialize: cs")
	// App Env
	env := os.Getenv("APP_ENV")
	switch env {
	case "dev", "develop":
		{
			INTERVAL = 1 * time.Second
		}

	case "test":
		{
			INTERVAL = 1 * time.Second

		}
	} // end switch
}

type listener interface {
	Receive(Configuration) error // client API, caller to do the implement.
}

type DefaultListener struct {
	Key           string
	DurationInSec int
	md5Human      string                    // md5 of old value
	Receive       func(Configuration) error // let the Caller register the func
}

func AddListener(li *DefaultListener) {
	commander.mu.Lock()
	defer commander.mu.Unlock()
	k := (*li).Key
	if _, existed := listeners[k]; !existed {
		listeners[k] = li
	}
	if !commander.running {
		go run()
	}
}

func removeListener(key string) {
	commander.mu.Lock()
	defer commander.mu.Unlock()
	delete(listeners, key)
}

// ConfigurationServer run
func run() {
	commander.running = true
	for {
		update()
		time.Sleep(10 * time.Second)
	}
}

func update() {
	for _, v := range listeners {
		cnf, err := Get(v.Key)
		if err != nil {
			log.Errorf("Cause' it can not get expected configuration. Error: %s.", err)
			// Do not panic
			continue
		}
		if f := v.Receive; f != nil {
			if cnf != nil {
				k := cnf.GetKey()
				if k == "" {
					// Having been deleted from CS , it affected the client.
					v.Receive(cnf)
					v.md5Human = ""
					removeListener(v.Key)
				} else {
					if cnf.GetMd5Human() != v.md5Human {
						v.md5Human = cnf.GetMd5Human()
						v.Receive(cnf)
					}
				}
			}
		} else {
			log.Error("Listener has no implemented function.", v)
		}
	}
}

type Configuration interface {
	GetKey() string
	GetValue() interface{} // string OR [](*SimpleConfiguration || *CombinedConfiguration)
	GetMd5Human() string
}

type SimpleConfiguration struct {
	Key      string `json:"key"`
	Md5Human string `json:"md5"`
	Value    string `json:"value"`
}

type CombinedConfiguration struct {
	Key      string      `json:"key"`
	Md5Human string      `json:"md5"`
	Value    interface{} `json:"value"` // sorted slice. sorted by Key.
	// Value:   []( *SimpleConfiguration||*CombinedConfiguration )
}

// By is the type of a "less" function that defines the ordering of its Configuration arguments.
type By func(cnf1, cnf2 interface{}) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(configurations []interface{}) {
	cnfS := &configurationSorter{
		configurations: configurations,
		by:             by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(cnfS)
}

// configurationSorter joins a By function and a slice of Configurations to be sorted.
type configurationSorter struct {
	configurations []interface{}
	by             By // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *configurationSorter) Len() int {
	return len(s.configurations)
}

// Swap is part of sort.Interface.
func (s *configurationSorter) Swap(i, j int) {
	s.configurations[i], s.configurations[j] = s.configurations[j], s.configurations[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *configurationSorter) Less(i, j int) bool {
	return s.by(s.configurations[i], s.configurations[j])
}

func (cnf *SimpleConfiguration) GetKey() string {
	return cnf.Key
}

func (cnf *SimpleConfiguration) GetValue() interface{} {
	return cnf.Value
}

func (cnf *SimpleConfiguration) GetMd5Human() string {
	if cnf.Md5Human != "" {
		return cnf.Md5Human
	}
	if cnf.Value == "" {
		cnf.Md5Human = EMPTY_STR_MD5
		return EMPTY_STR_MD5
	}
	data := []byte(cnf.Value)
	bs := md5.Sum(data)
	md5H := hex.EncodeToString(bs[:])
	cnf.Md5Human = md5H
	return md5H
}

// return Configuration , one of children : *SimpleConfiguration || *CombinedConfiguration
func (cnf *CombinedConfiguration) Child(key string) interface{} {
	//TODO Performance by using HashMap
	if vv, ok := (cnf.Value).([]interface{}); ok {
		for _, v := range vv {
			switch child := (v).(type) {
			case *SimpleConfiguration:
				{
					if child.Key == key {
						return child
					}
				}
			case *CombinedConfiguration:
				{
					if child.Key == key {
						return child
					}
				}
			default:
				{
					return nil
				}
			} // end switch
		} // end loop
	} else {
		log.Error("Have no children!!!")
		return nil
	}
	return nil
}

func (cnf *CombinedConfiguration) GetKey() string {
	return cnf.Key
}

func (cnf *CombinedConfiguration) GetValue() interface{} {
	return cnf.Value
}

// compute the all subConfigurations' md5
func (cnf *CombinedConfiguration) GetMd5Human() string {
	if cnf.Md5Human != "" && len(cnf.Md5Human) == MD5_CHARS_LEN {
		return cnf.Md5Human
	}
	hash := md5.New()
	// Firstly sort the cnf.Value ( []interface{} ) , Then compute Md5
	if val, ok := (cnf.Value).([]interface{}); ok {
		if len(val) <= 0 {
			cnf.Md5Human = EMPTY_STR_MD5
			return EMPTY_STR_MD5
		}
		By(keyCompareFunc).Sort(val)
		for _, v := range val {
			switch vv := v.(type) {
			case *SimpleConfiguration:
				{
					io.WriteString(hash, vv.GetMd5Human())
				}
			case *CombinedConfiguration:
				{
					io.WriteString(hash, vv.GetMd5Human())
				}
			default:
				{
					log.Info("Wrong element in ( CombinedConfiguration.Value ).")
					return ""
				}
			} // end switch
		} // end loop
	} else {
		log.Info("The CombinedConfiguration.Value is wrong type. Expected: []interface{} .")
		return ""
	}
	bs := hash.Sum(nil)
	md5H := hex.EncodeToString(bs[:])
	cnf.Md5Human = md5H
	return md5H
}

// Used to avoid recursion in UnmarshalJSON below.
type _simple SimpleConfiguration
type _combined CombinedConfiguration

func (cnf *CombinedConfiguration) UnmarshalJSON(bs []byte) (err error) {
	m := make(map[string]interface{})
	if err = json.Unmarshal(bs, &m); err == nil {
		cnf.Key = m["key"].(string)
		cnf.Md5Human = m["md5"].(string)
		if m["value"] != nil {
			vv := m["value"].([]interface{})
			cnf.Value, err = convertValue(vv)
		}
		return
	}
	log.Error(err)
	return
}

// return []( *Simple || *Combined )
func convertValue(val []interface{}) (array []interface{}, err error) {
	if len(val) == 0 {
		return
	}
	array = make([]interface{}, 0)
	for _, one := range val {
		// IF one element decoded type == *Simple , THEN directly assign the value
		// IF one element decoded type != *Combined, THEN recurs calling method
		m, ok := one.(map[string]interface{})
		if !ok {
			return nil, &(json.UnmarshalFieldError{})
		}
		key, md5 := m["key"].(string), m["md5"].(string)
		key = strings.TrimLeft(key, " ")
		key = strings.TrimRight(key, " ")
		var i interface{}
		if vv, ok := m["value"].(string); ok {
			vv = strings.TrimLeft(vv, " ")
			vv = strings.TrimRight(vv, " ")
			// v is simple
			i = &SimpleConfiguration{Key: key, Md5Human: md5, Value: vv}
		} else {
			if m["value"] != nil {
				vv, err := convertValue(m["value"].([]interface{}))
				if err != nil {
					log.Error("When convertValue, Error:", err)
					return nil, err
				}
				i = &CombinedConfiguration{Key: key, Md5Human: md5, Value: vv}
			}
		}
		if i != nil {
			array = append(array, interface{}(i))
		}
	}
	return
}
