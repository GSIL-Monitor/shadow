package cs

import (
	"log"
	"reflect"
	"testing"
	"time"
)

func init() {
	log.Println("Initialize: cs_test")
}

// (Configuration) GetMd5Human() string
func TestMd5Human(t *testing.T) {
	combMd5 := "7b54ff1510e6001ed798a580cc35931d"
	simple1 := &SimpleConfiguration{
		Key:      "a_simple",
		Md5Human: "77529b68d130d125ff07a534a7b69adb",
		Value:    "1|",
	}
	simple2 := &SimpleConfiguration{
		Key:      "b_simple",
		Md5Human: "43b864bde60564a9abad0598b1d5d203",
		Value:    "2|",
	}
	simple3 := &SimpleConfiguration{
		Key:      "c_simple",
		Md5Human: "eec250c1ce93e1977f730c44dd8e5e4d",
		Value:    "3|",
	}
	val := make([]interface{}, 0)
	val = append(val, interface{}(simple1))
	val = append(val, interface{}(simple2))
	val = append(val, interface{}(simple3))
	topComb := &CombinedConfiguration{
		Key:      "top",
		Md5Human: combMd5,
		Value:    val,
	}
	json := `
        {
            "key": "top",
            "md5": "",
            "value": [
                {
                    "key": "a_simple",
                    "md5": "77529b68d130d125ff07a534a7b69adb",
                    "value": "1|"
                },
                {
                    "key": "c_simple",
                    "md5": "eec250c1ce93e1977f730c44dd8e5e4d",
                    "value": "3|"
                },
                {
                    "key": "b_simple",
                    "md5": "43b864bde60564a9abad0598b1d5d203",
                    "value": "2|"
                } 
            ]
        }
    `
	actualTopComb, err := decodeByStr(json)
	if err != nil {
		t.Errorf("%s", err)
	}
	vv, ok := actualTopComb.(*CombinedConfiguration)
	if !ok {
		t.Error("Not ok.")
	}
	// decode , into Md5Human
	if vv.Md5Human != "" {
		t.Error("Not expected CombinedConfiguration: ", vv)
	}
	newMd5 := vv.GetMd5Human()
	if combMd5 != newMd5 {
		t.Error("Not equal. Excepted: ", combMd5, " . Actual Get: ", newMd5)
	}
	if topComb.Md5Human != vv.Md5Human {
		t.Error("Not equal. Excepted: ", topComb.Md5Human, " . Actual Get: ", vv.Md5Human)
	}
	if topComb.GetMd5Human() != vv.GetMd5Human() {
		t.Error("Not equal. Excepted GetMd5Human():  ", topComb.GetMd5Human(), " . Actual GetMd5Human(): ", vv.GetMd5Human())
	}
}

func TestConfigurationSort(t *testing.T) {
	simple1 := &SimpleConfiguration{
		Key:      "a_simple",
		Md5Human: "77529b68d130d125ff07a534a7b69adb",
		Value:    "1|",
	}
	simple2 := &SimpleConfiguration{
		Key:      "b_simple",
		Md5Human: "43b864bde60564a9abad0598b1d5d203",
		Value:    "2|",
	}
	simple3 := &SimpleConfiguration{
		Key:      "k_simple",
		Md5Human: "eec250c1ce93e1977f730c44dd8e5e4d",
		Value:    "3|",
	}
	simple4 := &SimpleConfiguration{
		Key:      "z_simple",
		Md5Human: "eec250c1ce93e1977f730c44dd8e5e4d",
		Value:    "4|",
	}
	simples := make([]interface{}, 0)
	simples = append(simples, interface{}(simple1))
	simples = append(simples, interface{}(simple3))
	simples = append(simples, interface{}(simple2))
	simples = append(simples, interface{}(simple4))
	assertSort(simples, t)

	ss1 := make([]interface{}, 0)
	ss1 = append(ss1, interface{}(simple4))
	ss1 = append(ss1, interface{}(simple3))
	ss1 = append(ss1, interface{}(simple1))
	ss1 = append(ss1, interface{}(simple2))
	ss1 = append(ss1, interface{}(simple3))
	assertSort(ss1, t)

}

// func (cnf *CombinedConfiguration) Child(key string) interface{}
func TestChild(t *testing.T) {
	combMd5 := "7b54ff1510e6001ed798a580cc35931d"
	simple1 := &SimpleConfiguration{
		Key:      "a_simple",
		Md5Human: "77529b68d130d125ff07a534a7b69adb",
		Value:    "1|",
	}
	simple2 := &SimpleConfiguration{
		Key:      "b_simple",
		Md5Human: "43b864bde60564a9abad0598b1d5d203",
		Value:    "2|",
	}
	simple3 := &SimpleConfiguration{
		Key:      "c_simple",
		Md5Human: "eec250c1ce93e1977f730c44dd8e5e4d",
		Value:    "3|",
	}
	val := make([]interface{}, 0)
	val = append(val, interface{}(simple1))
	val = append(val, interface{}(simple2))
	val = append(val, interface{}(simple3))
	topComb := &CombinedConfiguration{
		Key:      "top",
		Md5Human: combMd5,
		Value:    val,
	}
	key := "b_simple"
	child := topComb.Child(key)
	if !reflect.DeepEqual(simple2, child) {
		t.Errorf("Not equal. Expected:%q, Actual:%q\n", simple2, child)
	}
	key = "not_exist"
	child = topComb.Child(key)
	if child != nil {
		t.Errorf("Not expected. Actual:%q\n", child)
	}
}

// func AddListener(li *DefaultListener)
func TestAddListener(t *testing.T) {
	recFunc := func(Configuration) error {
		log.Println("I am a Listener.")
		return nil
	}
	li := &DefaultListener{
		Key:           "TestAddListener",
		DurationInSec: 1,
		Receive:       recFunc,
	}
	oldLen := len(listeners)
	AddListener(li)
	if len(listeners) != (oldLen + 1) {
		t.Error("Add listener fail.")
	}
	// Delay 2 secs, to judge
	time.Sleep(2 * time.Second)
	if !commander.running {
		t.Error("Not expected running status.", commander)
	}

}

// func RemoveListener(key string)
func TestRemoveListener(t *testing.T) {
	recFunc := func(Configuration) error {
		log.Println("I am a Listener.")
		return nil
	}
	li := &DefaultListener{
		Key:           "TestAddListener",
		DurationInSec: 1,
		Receive:       recFunc,
	}
	AddListener(li)
	removeListener(li.Key)
	_, ok := listeners[li.Key]
	if ok {
		t.Error(li, " exists in the SDK.")
	}
}

// func update()
func TestUpdate(t *testing.T) {
	changed := false
	key := "env_test_app_Shadow"
	recFunc := func(Configuration) error {
		log.Println("I am a Listener. Prepare to update...")
		changed = true
		return nil
	}
	li := &DefaultListener{
		Key:           key,
		DurationInSec: 1,
		Receive:       recFunc,
	}
	AddListener(li)
	if env == "test" {
		cnf, err := Get(key)
		if err == nil {
			if cnf.GetKey() != key {
				t.Error(cnf)
			}
		} else {
			t.Errorf("Getting ", key, "from CS has one error.")
		}
		if cnf.GetValue().(string) != "Shadow.Ring.test_ring" {
			t.Error("Not exptected configuration of CS.", cnf)
		}
		start := time.Now()
		end := start.Add(31 * time.Second)
		for {
			if !changed {
				time.Sleep(1 * time.Second)
			}
			now := time.Now()
			if changed || now.After(end) {
				break
			}
		}
		if !changed {
			t.Error("Listener does not receive changed info....")
		}
	}
}

func assertSort(simples []interface{}, t *testing.T) {
	By(keyCompareFunc).Sort(simples)
	if len(simples) > 1 {
		for i, s := range simples {
			if i == 0 {
				continue
			}
			pre, cur := simples[i].(*SimpleConfiguration), s.(*SimpleConfiguration)
			if !(pre.GetKey() <= cur.GetKey()) {
				t.Error("Not expected sorted array.", simples)
			}
		}

	}
}
