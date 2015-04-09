// This ccess Furion Serverfile as the client helper.
package cs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkgs
)

var (
	ErrEmpty = errors.New("No such concrete element.")
	ErrHttp  = errors.New("Http Access errors")
)
var (
	LocalDir      = os.Getenv("HOME") + "/.furion/local/"
	RemoteDir     = os.Getenv("HOME") + "/.furion/cache/" // To cache the remote data implemented by Files
	furionBaseUrl = ""
	localClient   = local{locks: make(map[string]sync.Mutex)}
	remoteClient  = remote{locks: make(map[string]sync.Mutex)}
)

const (
	DEFAULT_MODE_PERM os.FileMode = 0755
)

func init() {
	log.Info("Initialize: client")
	// App Env
	env := os.Getenv("APP_ENV")
	switch env {
	case "prod", "production":
		{
			furionBaseUrl = "http://furion.service.mogujie.org:8080/config/?key="
		}
	case "test":
		{
			furionBaseUrl = "http://192.168.2.120:8080/config/?key="
		}
	case "dev":
		{
			furionBaseUrl = "http://127.0.0.1:8080/config/?key="
		}
	default:
		{
			furionBaseUrl = "http://127.0.0.1:8080/config/?key="
		}
	}
	// create dir
	os.MkdirAll(LocalDir, DEFAULT_MODE_PERM)
	os.MkdirAll(RemoteDir, DEFAULT_MODE_PERM)
}

type client interface {
	get(key string) (Configuration, error)
	put(key, Configuration string) error
}

type local struct {
	sync.RWMutex
	locks map[string]sync.Mutex
}

type remote struct {
	sync.RWMutex
	locks map[string]sync.Mutex
}

func Get(key string) (Configuration, error) {
	val, err := localClient.get(key)
	if err != nil {
		return remoteClient.get(key)
	}
	return val, err
}

// CS data from local machine
func (local *local) get(key string) (Configuration, error) {
	path := filepath.Join(LocalDir, key)
	content, err := readFile(path)
	if err != nil {
		return nil, err
	}
	return decodeByStr(content)
}

// CS data from remote server
// value from the Furion by RESTful API, read Cached data after HTTP errors
func (remote *remote) get(key string) (Configuration, error) {
	url := furionBaseUrl + key
	// locks map
	remote.RLock()
	lock, ok := remote.locks[key]
	remote.RUnlock()
	// locks map
	remote.Lock()
	if !ok {
		lock = sync.Mutex{}
		remote.locks[key] = lock
	}
	remote.Unlock()
	//
	remote.RLock()
	defer remote.RUnlock()
	//lock one key
	lock.Lock()
	defer lock.Unlock()
	path := filepath.Join(RemoteDir, key)

	var httLogicalOK bool
	resp, err := http.Get(url)
	if err == nil {
		// handle , http status code
		code := resp.StatusCode
		httLogicalOK = (code >= http.StatusOK && code < http.StatusMultipleChoices)
		if !httLogicalOK {
			info := fmt.Sprintf("Remote server not ok [200,300) ,that is %d", code)
			err = errors.New(info)
		}
	}
	if err != nil {
		// read Cached data after HTTP errors
		log.Errorf("Try to read from Cache! When access Furion Server,one error occurs during handling the request. Error:%q", err)
		content, err := readFile(path)
		if err == nil && len(content) > 0 {
			cnf, err := decodeByStr(content)
			if err != nil {
				deleteFile(path)
			}
			return cnf, err
		}
		if err != nil {
			log.Errorf("File: %s, Error: %s", path, err)
		}
	}
	if httLogicalOK {
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		// read Cached data after HTTP errors
		if err != nil {
			log.Errorf("Try to read from Cache! When access Furion Server,one error occurs during handling the resp-body. Error:%q", err)
			content, err := readFile(path)
			cnf, err := decodeByStr(content)
			if err != nil {
				deleteFile(path)
			}
			return cnf, err
		}
		content := string(body)
		// over write cache
		err = overwriteFile(path, content)
		if err != nil {
			// delete cache && logging
			deleteFile(path)
			log.Errorf("When over write file has one error: %s", err)
		}
		return decode(body)
	}
	return nil, err
}

func readFile(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err == nil {
		return string(buf), nil
	}
	return "", err
}

func overwriteFile(filename, content string) (err error) {
	err = ioutil.WriteFile(filename, []byte(content), DEFAULT_MODE_PERM)
	if err != nil {
		log.Errorf("File: %s, Error: %s", filename, err)
	}
	return
}

// pure file. not dir
func deleteFile(filename string) {
	b := IsFile(filename)
	if b {
		os.Remove(filename)
	}
}

func IsFile(filename string) bool {
	f, err := os.Stat(filename)
	if err != nil {
		log.Errorf("File: %s, Error: %s", filename, err)
		return false
	}
	return !f.IsDir()
}

// return *SimpleConfiguration or *CombinedConfiguration
func decodeByStr(data string) (Configuration, error) {
	return decode([]byte(data))
}

// return *SimpleConfiguration or *CombinedConfiguration
func decode(bs []byte) (cnf Configuration, err error) {
	if cnf, err = decodeSimple(bs); err == nil {
		return
	}
	if cnf, err = decodeCombined(bs); err == nil {
		return
	}
	if err != nil {
		log.Error("Neither simple nor combined! Check the http resp body!")
	}
	return nil, err
}

// return *SimpleConfiguration
func decodeSimple(bs []byte) (cnf Configuration, err error) {
	cnf = new(SimpleConfiguration)
	r := bytes.NewReader(bs)
	err = json.NewDecoder(r).Decode(cnf)
	if err == nil {
		c := cnf.(*SimpleConfiguration)
		vv := c.Value
		vv = strings.TrimLeft(vv, " ")
		vv = strings.TrimRight(vv, " ")
		c.Value = vv
		kk := c.Key
		kk = strings.TrimLeft(kk, " ")
		kk = strings.TrimRight(kk, " ")
		c.Key = kk
		return
	}
	log.Error(err)
	return
}

// return *CombinedConfiguration
func decodeCombined(bs []byte) (Configuration, error) {
	cnf := new(CombinedConfiguration)
	err := json.Unmarshal(bs, &cnf)
	if err != nil {
		log.Error(err)
	}
	return cnf, err
}
