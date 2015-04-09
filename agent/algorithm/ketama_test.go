package algorithm

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	cases := []struct {
		in   string
		want uint32
	}{
		{"26", uint32(2309692575)},
		{"1404", uint32(2175855552)},
		{"4177", uint32(2010484542)},
		{"9315", uint32(303768557)},
		{"14745", uint32(3947945572)},
		{"105106", uint32(1628205094)},
		{"355107", uint32(680118014)},
	}

	for _, c := range cases {
		got := Hash(c.in)
		if got != c.want {
			t.Errorf("Hash method wrong. Input: %s, Expected: %d, Actual: %d", c.in, c.want, got)
		}
	}
}

func newInstances() map[string]uint32 {
	instances := make(map[string]uint32)
	instances["10.0.1.1"] = 2
	instances["10.0.1.2"] = 2
	instances["10.0.1.3"] = 2
	instances["10.0.1.4"] = 2
	instances["10.0.1.5"] = 2

	instances["10.0.1.11"] = 2
	instances["10.0.1.12"] = 2
	instances["10.0.1.13"] = 2
	instances["10.0.1.14"] = 2
	instances["10.0.1.15"] = 2

	instances["10.1.1.1"] = 1
	return instances
}

func TestNewRing(t *testing.T) {
	instances := newInstances()
	// VNs: 2048, totalWeight: 21, instances: 11
	r, err := NewRing(instances)

	if err != nil {
		t.Errorf("Error: %v \n", err)
	}

	totalWeight := uint32(0)
	for _, w := range instances {
		totalWeight += w
	}

	fmt.Printf("Actual Total Weight: %d\n", r.totalWeight)
	if r.totalWeight != totalWeight {
		t.Errorf("%d\n", r.zoomFactor)
	}

	fmt.Printf("Get Virtual Nodes: %d\n", len(r.nodes))
	if len(r.nodes) < int(MIN_VIRTUAL_NODES) {
		t.Errorf("Virtual Nodes too less, so the ring is too small !")
	}

	if r.zoomFactor < MIN_ZOOM_FACTOR {
		t.Errorf("ZoomFactor. Actual: %d , expected MIN_ZOOM_FACTOR: %d", r.zoomFactor, MIN_ZOOM_FACTOR)
	}
}

func TestGetMaxKey(t *testing.T) {
	instances := newInstances()
	r, _ := NewRing(instances)
	max := uint32(0)
	currentHash := uint32(0)
	for _, node := range r.nodes {
		if node.hash >= max {
			max = node.hash
		}
		if currentHash > node.hash {
			t.Errorf("Virtual Nodes do not sort...")
		}
	}
	c := struct {
		in, want uint32
	}{
		in:   r.getMaxKey(),
		want: max,
	}

	if c.in != c.want {
		t.Errorf("actual: %d, excepted: %d", c.in, c.want)
	}
}

func TestAddInstance(t *testing.T) {
	instances := newInstances()
	r, _ := NewRing(instances)
	oldInstancesCnt := len(instances)
	oldLen := len(r.nodes)
	addr := "10.2.1.1"
	w := uint32(1)
	r.AddInstance(addr, w)
	currentLen := len(r.nodes)
	instancesCnt := len(r.instances)
	if oldLen >= currentLen {
		t.Errorf("Except to add one instance into the ring! But virtual nodes don't increase!")
	}
	if currentLen != r.length {
		t.Errorf("Virtual Nodes is not consistent!")
	}
	if oldInstancesCnt+1 != instancesCnt {
		t.Errorf("Except to add one instance into the ring!")
	}
	if int(r.zoomFactor)+oldLen != currentLen {
		t.Errorf("The ring does not match the zoom factor!")
	}
}

func TestAddInstances(t *testing.T) {
	instances := newInstances()
	r, _ := NewRing(instances)
	oldInstancesCnt := len(instances)
	oldLen := len(r.nodes)
	adds := make(map[string]uint32)
	adds["10.2.1.1"] = 1
	adds["10.2.1.2"] = 1
	r.AddInstances(adds)

	currentLen := len(r.nodes)
	instancesCnt := len(r.instances)
	if oldLen >= currentLen {
		t.Errorf("Except to add one instance into the ring! But virtual nodes don't increase!")
	}
	if currentLen != r.length {
		t.Errorf("Virtual Nodes is not consistent!")
	}
	if oldInstancesCnt+2 != instancesCnt {
		t.Errorf("Except to add one instance into the ring!")
	}
}

func newEfficientInstances() map[string]uint32 {
	instances := make(map[string]uint32)
	instances["10.0.1.1"] = 1
	instances["10.0.1.2"] = 1
	instances["10.0.1.3"] = 1
	instances["10.0.1.4"] = 1
	instances["10.0.1.5"] = 1

	instances["10.0.1.11"] = 1
	instances["10.0.1.12"] = 1
	instances["10.0.1.13"] = 1
	instances["10.0.1.14"] = 1
	instances["10.0.1.15"] = 1

	instances["10.0.2.1"] = 1
	instances["10.0.2.2"] = 1
	instances["10.0.2.3"] = 1

	return instances
}

func TestRemoveInstance(t *testing.T) {
	instances := newEfficientInstances()
	r, err := NewRing(instances)
	if err != nil {
		t.Errorf("%q", err)
	}
	oldInstancesCnt := len(instances)
	oldLen := len(r.nodes)

	r.RemoveInstance("10.0.2.1")
	currentLen := len(r.nodes)
	instancesCnt := len(r.instances)
	if oldLen <= currentLen {
		t.Errorf("Except to remove one instance from the ring.But the virtual nodes do not decrease.")
	}
	if currentLen != r.length {
		t.Errorf("Virtual Nodes is not consistent!")
	}
	if oldInstancesCnt-1 != instancesCnt {
		t.Errorf("Except to remove one instance from the ring! Physical size, old: %d , new: %d", oldInstancesCnt, instancesCnt)
	}
}

func TestPickMaster(t *testing.T) {
	instances := newEfficientInstances()
	r, err := NewRing(instances)
	if err != nil {
		t.Errorf("Have one error:%s", err)
		return
	}
	key := "cpc_booking_u-123_page-1"
	addr := r.PickMaster(key)
	_, ok := instances[addr]
	if !ok {
		t.Errorf("Instance does not exist in the ring! Address: %s", addr)
	}
}
