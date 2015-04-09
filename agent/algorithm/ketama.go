package algorithm

// Ketama Hash return 0 ~ (2^32 - 1) .
// Inner uses lower letter if not unix_socket .
// Distributed system does not use unix_socket .
// Hash Value >= 0 , unsigned value
// Try its best to defined Unsigned Value .

// Each instance default weight is 1 (recommand), that must be >= 1 .

import (
	"crypto/sha1"
	"errors"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	// third pkgs
	log "github.com/cihub/seelog"
)

const (
	MIN_VIRTUAL_NODES  = uint32(2048)
	MAX_TOTAL_WEIGHT   = uint32(math.MaxUint32 / MIN_VIRTUAL_NODES)
	MIN_ZOOM_FACTOR    = uint32(100) // Theoretical ZOOM_FACTOR : 100~200
	MAX_ZOOM_FACTOR    = uint32(200) // Theoretical ZOOM_FACTOR : 100~200
	SAFE_NUM_INSTANCES = uint32(MIN_VIRTUAL_NODES/MAX_ZOOM_FACTOR + 1)
)

var (
	ErrTotalWeightOverflow = errors.New("Ketama: the total weight is overflow! Please minimize it!")
	ErrNoServer            = errors.New("No valid server definitions found.")
	ErrInvalidWeight       = errors.New("Please set weight >= 1.")
	ErrNodeNotFound        = errors.New("No node found in the existed ring.")
	ErrExistedNode         = errors.New("Your input new node actually exists in the Ring!")
)

// domain modeling . key : addr , value : weight
type instances map[string]uint32

// Virtual Node in Ketama
type node struct {
	address string // ip:port  or  hostname:port
	hash    uint32 // in the Ring
}

type nodes []node

func (r nodes) Len() int           { return len(r) }
func (r nodes) Less(i, j int) bool { return r[i].hash < r[j].hash }
func (r nodes) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r nodes) Sort()              { sort.Sort(r) }

type HashRing struct {
	instances   instances // Physical Nodes
	nodes       nodes     // sorted nodes
	zoomFactor  uint32    // do not modify after initialization
	length      int       // > uint32 nodes is horrible big pool
	totalWeight uint32
	// mutex
	mu sync.Mutex
}

func Hash(in string) uint32 {
	hash := sha1.New()
	hash.Write([]byte(in))
	digest := hash.Sum(nil)
	v := uint32(digest[19]) | uint32(digest[18])<<8 | uint32(digest[17])<<16 | uint32(digest[16])<<24
	return v
}

// NewRing is the online min Ring with Virtual Nodes
func NewRing(instances map[string]uint32) (ring *HashRing, err error) {
	s := len(instances)
	if s <= 0 {
		log.Error("No servers so that can not execute NewRing")
		return nil, ErrNoServer
	}

	if s < int(SAFE_NUM_INSTANCES) {
		log.Info("Your instances is not enough! Please increase the number of instances in one ring to ", SAFE_NUM_INSTANCES)
	}

	ring = new(HashRing)

	totalWeight := uint32(0)
	for _, w := range instances {
		// weight  >= 1
		if w <= 0 {
			log.Error("Weight is invalid, please check the configuration")
			return nil, ErrInvalidWeight
		}
		totalWeight += uint32(w)
	}

	if totalWeight >= MAX_TOTAL_WEIGHT {
		log.Error("It is horrible, too big total weight.")
		return nil, ErrTotalWeightOverflow
	}

	quotient := uint32(MIN_VIRTUAL_NODES / totalWeight)
	zoomFactor := uint32(0)
	if MIN_VIRTUAL_NODES%totalWeight == 0 {
		zoomFactor = quotient
	} else {
		zoomFactor = quotient + 1
	}

	if zoomFactor < MIN_ZOOM_FACTOR {
		log.Infof("Have %d instances and totalWeight: %d,but zoomFactor: %d is too small. Now change to MIN_ZOOM_FACTOR explicitly", s, totalWeight, zoomFactor)
		zoomFactor = MIN_ZOOM_FACTOR
	}

	// the format of KeyForNode : 10.15.0.200:6379-0 , 10.15.0.200:6379-1
	for addr, w := range instances {
		vCnt := int(uint32(zoomFactor) * w)
		for i := 0; i < vCnt; i++ {
			key := addr + "-" + strconv.Itoa(i)
			hashing := Hash(key)
			// one virtual node
			vn := &node{address: addr, hash: hashing}
			ring.nodes = append(ring.nodes, *vn)
		}
	}
	// initialize
	ring.nodes.Sort()
	ring.instances = instances
	ring.length = len(ring.nodes)
	ring.totalWeight = uint32(totalWeight)
	ring.zoomFactor = uint32(zoomFactor)
	// post check to tip
	if ring.length >= int(math.MaxUint32/4*3) {
		log.Infof("Too large ring, has %d virtual nodes", ring.length)
	}
	return
}

// virtual node , max hash value
func (ring *HashRing) getMaxKey() uint32 {
	return ring.nodes[ring.length-1].hash
}

// situation : Just add one new instance dynamically
func (ring *HashRing) AddInstance(addr string, weight uint32) error {
	// instance should not be in the existed ring
	// copy from the existed ring
	newNodes := make(nodes, ring.length, ring.length+int(weight*ring.zoomFactor))
	copy(newNodes, ring.nodes)
	for _, n := range newNodes {
		if strings.EqualFold(addr, n.address) {
			log.Errorf("The %s add exists in runtime", addr)
			return ErrExistedNode
		}
	}
	vCnt := int(ring.zoomFactor * weight)
	for i := 0; i < vCnt; i++ {
		key := addr + "-" + strconv.Itoa(i)
		hashing := Hash(key)
		// one virtual node
		vn := &node{address: addr, hash: hashing}
		newNodes = append(newNodes, *vn)
	}
	newNodes.Sort()

	ring.mu.Lock()
	defer ring.mu.Unlock()
	ring.instances[addr] = weight
	ring.nodes = newNodes
	ring.length = len(ring.nodes)
	ring.totalWeight = uint32(ring.totalWeight + weight)
	return nil
}

func (ring *HashRing) AddInstances(instances map[string]uint32) {
	ring.mu.Lock()
	defer ring.mu.Unlock()
	// nodes should not be in the existed ring,
	// so do filter the existed instance
	var adds = make(map[string]uint32)
	for name, w := range instances {
		_, ok := ring.instances[name]
		if !ok {
			// node not in instances , will append
			adds[name] = w
		}
	}
	if len(adds) <= 0 {
		log.Error("Have no address from instances map")
		return
	}
	// copy from the existed ring
	newNodes := make(nodes, ring.length)
	copy(newNodes, ring.nodes)

	appendWeight := uint32(0)
	for n, w := range adds {
		appendWeight += w
		vCnt := int(ring.zoomFactor * w)
		for i := 0; i < vCnt; i++ {
			key := n + "-" + strconv.Itoa(i)
			hashing := Hash(key)
			// one virtual node
			vn := &node{address: n, hash: hashing}
			newNodes = append(newNodes, *vn)
		}
		ring.instances[n] = w
	}
	if len(newNodes) <= len(ring.nodes) {
		log.Info("No more node need to be added")
		return
	}

	newNodes.Sort()
	ring.nodes = newNodes
	ring.length = len(ring.nodes)
	ring.totalWeight = uint32(ring.totalWeight + appendWeight)
}

func (ring *HashRing) RemoveInstance(addr string) error {
	ring.mu.Lock()
	defer ring.mu.Unlock()
	// copy-on-write
	newNodes := make(nodes, 0, ring.length)
	removeCount := uint32(0)
	for _, v := range ring.nodes {
		if !strings.EqualFold(addr, v.address) {
			newNodes = append(newNodes, v)
		} else {
			removeCount = removeCount + 1
		}
	}
	if len(newNodes) == len(ring.nodes) {
		log.Error("No found %s need to be removed", addr)
		return ErrNodeNotFound
	}

	weight := ring.instances[addr]
	ring.nodes = newNodes
	ring.length = len(ring.nodes)
	ring.totalWeight = ring.totalWeight - weight
	if ring.length < int(SAFE_NUM_INSTANCES) {
		log.Info("Virtual Nodes are not enough caused by you remove instances!")
	}
	delete(ring.instances, addr)
	return nil
}

func (ring *HashRing) PickMaster(key string) string {
	v := Hash(key)
	i := sort.Search(ring.length, func(i int) bool { return ring.nodes[i].hash >= v })
	if i >= ring.length {
		i = 0
	}
	return ring.nodes[i].address
}
