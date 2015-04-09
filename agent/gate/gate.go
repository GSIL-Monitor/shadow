// Package gate provides manager all the physical nodes and the virtual nodes.
// CRUD operations
// Contact configuration server && Do dynamic configs
package gate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkgs
	alg "shadow/agent/algorithm"
	cnfPkg "shadow/agent/cnf"
	"shadow/agent/cs"
	"shadow/agent/redis"
)

// ------------------------------------------
// - App cares about keys. CS above:
// - key : {AppName}__Shadow
// - value : {RingName}
// ------------------------------------------

// ------------------------------------------
// - Ring cares about keys. CS above:
// - key : Shadow.Ring.{RingName}
// - value : string({shadow.{RingName}.yaml})
// ------------------------------------------

const (
	APP_CARE_KEYS_FORMAT = "%s__Shadow"
	APP_CARE_RING_FORMAT = "Shadow.Ring.%s"
)

var (
	ErrAccessDenied      = errors.New("No access rights.")
	ErrNewRing           = errors.New("Fatal: during new ring!")
	ErrNoRing            = errors.New("Fatal: no ring!")
	localGate       Gate = Gate{
		apps:  make(map[string]string),
		rings: make(map[string]*ring),
	}
)

// expose API to higher layer
type GateActions interface {
	Locate(appName, key string) (conn redis.Connection, err error)
}

type Gate struct {
	apps  map[string]string // App_2_RingName
	rings map[string]*ring  // RingName_2_Ring. RingName_2_nil: when changing...
	mu    sync.RWMutex
	GateActions
}

// One physical ring consists of N redis(with conns-pool) in the deployment diagram.
type ring struct {
	// physical redis connections. Address_2_Pool
	pools map[string]*redis.Pool
	// ketama algorithm. include basic infos and virtual nodes
	hashRing *alg.HashRing
	// Mutex
	mu sync.RWMutex
}

func init() {
	log.Info("Initialize: gate")
	asyncGate()
}

// Prepare the environment for running application.
func asyncGate() {
	go func() {
		// static configuration
		initApps()
		// dynamic configuration
		initListeners()
		log.Info("Initialize listeners done")
	}()
	checkApps()
	checkRings()
}

func initApps() {
	// configuration
	apps := getApps()
	if len(apps) < 1 {
		log.Info("No apps. Why? Please check the configuration.")
		return
	}
	// A set of ring name . DataStruct: set
	for _, app := range apps {
		//TODO extension: ACL
		ringName, err := getRingName(app)
		if err != nil {
			log.Error(err)
		} else {
			localGate.mu.Lock()
			if ringName != "" {
				log.Info("Get one app: ", app)
				localGate.apps[app] = ringName
			} else {
				log.Info("The ringName is empty content from CS. AppName: ", app)
				// Access Denied
				delete(localGate.apps, app)
			}
			localGate.mu.Unlock()
		}
	}
}

// Do it right now. Add listeners so then app adds pools
func initListeners() {
	// add listeners to the CS SDK
	// all subscribed keys
	dst := make(map[string]string)
	for k, v := range localGate.apps {
		dst[k] = v
	}
	if len(dst) < 1 {
		log.Info("No ring name set! Please check all of the app!")
		return
	}
	for app, _ := range dst {
		addCaredKey(app)
	}

	// handle the common ring shared by multi apps
	subscribeRings := make(map[string]bool)
	for _, ringName := range localGate.apps {
		subscribeRings[ringName] = true
	}
	for ringName, _ := range subscribeRings {
		addCaredRing(ringName)
	}
} // End func

// daemon check local file for a list of app name
func checkApps() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			apps := getApps()
			if len(apps) < 1 {
				log.Info("Check apps in daemon, but no apps. Why? Please check the configuration.")
				continue
			}
			var ringName string
			var ok bool
			for _, app := range apps {
				addCaredKey(app)
				if ringName, ok = localGate.apps[app]; ok {
					time.Sleep(1 * time.Second)
					// existed ringName in Gate struct
					localGate.mu.RLock()
					_, ringOK := localGate.rings[ringName]
					localGate.mu.RUnlock()
					if !ringOK {
						// has new ring struct
						addCaredRing(ringName)
					}
				} else {
					// no ringName
					// has new app that need to subscribe ring_name
				}
			}
		}
	}()
}

// daemon check localGate to remove the redundant pools
func checkRings() {
	go func() {
		for {
			time.Sleep(10 * time.Second)
			doCheckRings()
		}
	}()
}

// primitive operation
func doCheckRings() {
	destroyRings := make(map[string]*ring)
	// App_2_RingName apps
	// RingName_2_Ring rings
	for ringNameInRings, ring := range localGate.rings {
		exist := false
	apps:
		for _, ringName := range localGate.apps {
			if ringNameInRings == ringName {
				exist = true
				break apps
			}
		}
		if !exist {
			destroyRings[ringNameInRings] = ring
		}
	}
	for ringName, r := range destroyRings {
		log.Info("Prepare to destroy ring: ", ringName)
		if r != nil {
			r.Destroy()
		}
		localGate.mu.Lock()
		delete(localGate.rings, ringName)
		localGate.mu.Unlock()
	}
}

// Let the Caller handle the error.
func Locate(appName, key string) (conn redis.Connection, err error) {
	localGate.mu.RLock()
	defer localGate.mu.RUnlock()
	ringName, ok := localGate.apps[appName]
	if !ok || ringName == "" {
		log.Errorf("%s can not access the ring.", appName)
		return nil, ErrAccessDenied
	}
	var r *ring
	for i := 0; i < 3 && r == nil; i++ {
		if r = localGate.rings[ringName]; r == nil {
			time.Sleep(1 * time.Millisecond)
		}
	}
	if r == nil {
		return nil, ErrNoRing
	}
	// Get redis.Connection-Pool by key
	addr := r.hashRing.PickMaster(key)
	if pool := r.pools[addr]; pool != nil {
		if conn, err = pool.Get(); err != nil {
			log.Errorf("AppName: %s, Error: %s", appName, err)
		}
	}
	return
}

func (gate *Gate) Close() error {
	gate.mu.Lock()
	defer gate.mu.Unlock()
	has := false
	for _, r := range gate.rings {
		if len(r.pools) > 0 {
			for addr, pool := range r.pools {
				err := pool.Close()
				if err != nil {
					has = true
					log.Errorf("When close address: %s, Error: %s", addr, err)
				}
			}
		}
	}
	if has {
		return errors.New("Have errors when close the gate.")
	}
	return nil
}

// A list of app name. From local file
func getApps() []string {
	f := filepath.Join(cnfPkg.CnfBasedir, "apps.yaml")
	cnf, err := ioutil.ReadFile(f)
	if err != nil {
		log.Errorf("Can not get a list of app name. The filepath: %s, Error: %s", f, err)
		return nil
	}
	return DecodeAppsYAML(cnf)
}

// return ring name
func getRingName(app string) (string, error) {
	// from CS
	key := fmt.Sprintf(APP_CARE_KEYS_FORMAT, app)
	cnf, err := cs.Get(key)
	if err != nil {
		log.Error("AppName: %s, Error: %s", app, err)
		return "", err
	}
	simple, ok := cnf.(*cs.SimpleConfiguration)
	if !ok {
		log.Error("Key: ", key, ", but the value is wrong, not a simple string.")
		return "", err
	}
	if simple.GetKey() == "" {
		// The configuration has been deleted from the CS server.
		return "", nil
	}
	// return only one RingName
	return simple.GetValue().(string), nil
}

// Address_2_Weight
func getInstancesFromCS(ring string) map[string]uint32 {
	// from CS
	key := fmt.Sprintf(APP_CARE_RING_FORMAT, ring)
	cnf, err := cs.Get(key)
	if err != nil {
		log.Error("Getting one ring from CS has one error. Error: ", err)
		return nil
	}
	simple, ok := cnf.(*cs.SimpleConfiguration)
	if !ok {
		log.Error("Key: ", key, ", get ring info , but the wrong value not a simple string.")
		return nil
	}
	if simple.GetKey() == "" {
		// The configuration has been deleted from the CS server.
		return nil
	}
	yaml := simple.GetValue().(string)
	return getInstances(ring, yaml)
}

func getInstances(ringName, yaml string) map[string]uint32 {
	ringYaml := DecodeRingYAMLByStr(ringName, yaml)
	if ringYaml == nil {
		log.Error("Get one ring occurs error! Actual: no RingYaml. Please check it!")
		return nil
	}
	return ringYaml.getInstances()
}

// instances: Address_2_Weight , return Address_2_Pool
func newPools(instances map[string]uint32) (map[string]*redis.Pool, error) {
	pools := make(map[string]*redis.Pool)
	// new pool will be put into the map
	for addr, _ := range instances {
		pools[addr] = newPool(addr, "")
	}
	return pools, nil
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   1,
		MaxIdle:     1,
		IdleTimeout: 600 * time.Second,
		Wait:        true,
		Dial: func() (redis.Connection, error) {
			// nc, err := net.DialTimeout("tcp", server, 1*time.Second)
			// if err != nil {
			// 	log.Info("Create one connection, a error occurs.", server)
			// 	return nil, err
			// }
			// tc, ok := nc.(*net.TCPConn)
			// if !ok {
			// 	return nil, errors.New("Not TCPConn")
			// }
			// if err := tc.SetKeepAlive(true); err != nil {
			// 	return nil, err
			// }
			// if err := tc.SetKeepAlivePeriod(29 * time.Minute); err != nil {
			// 	return nil, err
			// }
			// readTimeout := 1 * time.Second
			// writeTimeout := 1 * time.Second
			// c := redis.NewConn(tc, readTimeout, writeTimeout)

			timeout := 1 * time.Second
			c, err := redis.DialTimeout("tcp", server, timeout, timeout, timeout)
			if err != nil {
				log.Error("Create one connection, Error: ", server)
				return nil, err
			}
			if password != "" {
				log.Info("Need AUTH.")
			}
			return c, err
		},
	}
}

func (r *ring) Destroy() {
	for _, p := range r.pools {
		p.Close()
	}
	r.pools = nil
	r.hashRing = nil
}

func newRing(instances map[string]uint32) (r *ring, err error) {
	if len(instances) < 1 {
		return nil, ErrNoRing
	}
	pools, pErr := newPools(instances)
	hashRing, hErr := alg.NewRing(instances)
	if hErr == nil && pErr == nil {
		r = &ring{
			pools:    pools,
			hashRing: hashRing,
		}
		return
	} else {
		log.Error("New connection-pooling occurs: ", pErr)
		log.Error("New hash-ring occurs: ", hErr)
	}
	return r, ErrNewRing
}

// add listener
func addCaredKey(app string) {
	localGate.mu.Lock()
	defer localGate.mu.Unlock()
	if app == "" {
		log.Error("The AppName is empty!")
		return
	}
	key := fmt.Sprintf(APP_CARE_KEYS_FORMAT, app)
	// callback
	ringNameFunc := func(cnf cs.Configuration) (err error) {
		k := cnf.GetKey()
		newRingName := cnf.GetValue().(string)
		if key != k {
			log.Error("Having subscribed ring_name,Shadow got it. The expected key: ", key, ". Having got the key from CS: ", k)
		}
		if k == "" {
			// Having deleted from CS
			localGate.mu.Lock()
			delete(localGate.apps, app)
			localGate.mu.Unlock()
			return
		}
		localGate.mu.Lock()
		localGate.apps[app] = newRingName
		localGate.mu.Unlock()
		// IF got newRingName is empty, THEN no access privilege!
		if newRingName != "" {
			localGate.mu.RLock()
			_, ok := localGate.rings[newRingName]
			localGate.mu.RUnlock()
			// cascade
			if !ok {
				log.Info("Cascade add one new ring. The ring: ", newRingName)
				addCaredRing(newRingName)
			}
		}
		return
	} // end callback
	listener := &cs.DefaultListener{
		Key:           key,
		DurationInSec: 1,
		Receive:       ringNameFunc,
	}
	cs.AddListener(listener)
}

// add listener
func addCaredRing(ringName string) {
	localGate.mu.Lock()
	defer localGate.mu.Unlock()
	if ringName == "" {
		deleteOldRing(ringName)
		log.Error("The RingName is empty!")
		return
	}
	key := fmt.Sprintf(APP_CARE_RING_FORMAT, ringName)
	// callback
	ringFunc := func(cnf cs.Configuration) (err error) {
		k := cnf.GetKey()
		yaml := cnf.GetValue().(string)
		if key != k {
			log.Info("Having subscribe ring instances,Shadow got it. The expected key: ", key, ". Having got the key from CS: ", k)
		}
		if k == "" {
			// Having deleted from CS
			deleteOldRing(ringName)
			return
		}
		if yaml == "" {
			// one person fatal !
			deleteOldRing(ringName)
			return
		}
		// new ->( type ring struct ), to replace the old one or add new one
		log.Info("Prepare building the ring: ", ringName)
		instances := getInstances(ringName, yaml)
		newring, err := newRing(instances)
		if length := len(instances); err == nil && length > 0 {
			// store the old ring
			localGate.mu.RLock()
			old := localGate.rings[ringName]
			localGate.mu.RUnlock()
			if old != nil {
				go func() {
					old.Destroy()
				}()
			}
			// Replace/Add
			localGate.mu.Lock()
			localGate.rings[ringName] = newring
			localGate.mu.Unlock()
			log.Infof("The ring: %s, has %d instances", ringName, length)
		} else {
			// parse errors
			log.Info("The ring: ", ringName, ", has no instances!")
			if err != nil {
				log.Error("The ring: ", ringName, ". The error: ", err)
			}
		}
		return
	} // end callback
	listener := &cs.DefaultListener{
		Key:           key,
		DurationInSec: 1,
		Receive:       ringFunc,
	}
	cs.AddListener(listener)
}

func deleteOldRing(ringName string) {
	localGate.mu.RLock()
	old := localGate.rings[ringName]
	localGate.mu.RUnlock()
	if old != nil {
		localGate.mu.Lock()
		delete(localGate.rings, ringName)
		localGate.mu.Unlock()
		old.Destroy()
	}
}

func Shutdown() {
	localGate.Close()
}

// TODO
func Check() {
}
