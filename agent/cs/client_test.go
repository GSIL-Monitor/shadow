package cs

import (
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	//
)

var (
	env = os.Getenv("APP_ENV")
)

const ()

func init() {
	log.Println("client_test begin...")
	log.Println("APP_ENV:", env)
}

func TestGet(t *testing.T) {
	log.Println("BaseURL:", furionBaseUrl)
	key := "delete"
	if env == "dev" {
		path := filepath.Join(RemoteDir, key)
		deleteFile(path)
	}
	cnf, err := Get("delete")
	switch env {
	case "prod", "test":
		{
			if err != nil {
				t.Errorf("%q", err)
			}
			assertEmptyCS(t, cnf)
		}
	case "dev":
		{
			if err == nil {
				t.Errorf("Expect 404. Actual have no error, we can get Configuration.")
			}

		}
	} // end switch

	// SimpleConfiguration CombinedConfiguration
	if env == "test" {
		key := "env_test_app_Shadow"
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
	}
}

// func (local *local) get(key string) (Configuration, error)
func TestLocalGet(t *testing.T) {
	r := int(rand.Int31n(100))
	key := "TestLocalGet_" + strconv.Itoa(r)
	path := filepath.Join(LocalDir, key)

	content := `{"key":"TestLocalGet_","md5":"0bae7dd77b681aa2bc8c49d8e83a5865","value":"{\"failover\":1,\"methodTimeout\":null,\"timeout\":6000}"}`
	err := overwriteFile(path, content)
	if err != nil {
		t.Error(err)
	}
	defer deleteFile(path)

	cnf, err := localClient.get(key)
	if err != nil {
		t.Error(err)
	} else {
		v, ok := cnf.(*SimpleConfiguration)
		if !ok {
			t.Errorf("Type is wrong. %q", cnf)
		} else {
			if v.GetKey() != "TestLocalGet_" {
				t.Error("Having parsed configuration , it was wrong.")
			}
			val := v.GetValue().(string)
			if val == "" {
				t.Error("Having parsed configuration , it was wrong. Value: ", val)
			}
			if val != v.Value {
				t.Error("Having parsed configuration , it was wrong. Configuration: ", v)
			}
		}
	}
	key = "not_exist"
	_, err = localClient.get(key)
	if err == nil {
		t.Errorf("Expect error during no existed file.")
	}
}

// func (remote *remote) get(key string) (Configuration, error)
func TestRemoteGate(t *testing.T) {
	if env == "test" {
		key := "TestRemoteGate_DB"
		cnf, err := remoteClient.get(key)
		if err != nil {
			t.Error(err)
		}
		if cnf == nil {
			t.Error("No configuration")
		}
		if cnf.GetKey() != key {
			t.Error("Not expected key")
		}
		val, ok := cnf.GetValue().(string)
		if !ok {
			t.Error("From CS,", key, " . Unmarshal: ", cnf)
		}
		if val != "Shadow_Ring" {
			t.Error("Not expected configuration. Please check CS.")
		}
		// check cache
		path := filepath.Join(RemoteDir, key)
		if !IsFile(path) {
			t.Error("Not expected cache. Key:", key)
		}
		content, err := readFile(path)
		if err != nil {
			t.Error("SDK can not read cache from disk file. Path:", path)
		}
		actualCnf, err := decodeByStr(content)
		if !reflect.DeepEqual(cnf, actualCnf) {
			t.Errorf("Not equal. expected:%q, actual:%q\n", cnf, actualCnf)
		}
	}
}

// func readFile(filename string) (string, error)
func TestReadFile(t *testing.T) {
	r := int(rand.Int31n(100))
	key := "TestReadFile" + strconv.Itoa(r)
	path := filepath.Join(RemoteDir, key)
	content := `{"key":"cs_test_local","md5":"0bae7dd77b681aa2bc8c49d8e83a5865","value":"{\"failover\":1,\"methodTimeout\":null,\"timeout\":6000}"}`
	err := overwriteFile(path, content)
	if err != nil {
		t.Error(err)
	}
	defer deleteFile(path)
	if !IsFile(path) {
		t.Error("Not file")
	}
	actual, err := readFile(path)
	if content != actual {
		t.Error("Not equal conent of file.")
	}
	if err != nil {
		t.Error(err)
	}
}

// func overwriteFile(filename, content string) error
func TestOverwriteFile(t *testing.T) {
	r := int(rand.Int31n(100))
	key := "TestOverwriteFile" + strconv.Itoa(r)
	path := filepath.Join(RemoteDir, key)
	content := `TestOverwriteFile...`
	err := overwriteFile(path, content)
	if err != nil {
		t.Error(err)
	}
	defer deleteFile(path)
	if !IsFile(path) {
		t.Error("Not file")
	}
	// overwrite
	newContent := "TestOverwriteFile again..."
	err = overwriteFile(path, newContent)
	if err != nil {
		t.Error(err)
	}
	actual, err := readFile(path)
	if newContent != actual {
		t.Error("Not equal conent of file.")
	}
	if err != nil {
		t.Error(err)
	}
}

// func deleteFile(filename string)
func TestDeleteFile(t *testing.T) {
	content := "test...."
	key := "TestDeleteFile"
	path := filepath.Join(LocalDir, key)
	err := overwriteFile(path, content)
	if err != nil {
		t.Error(err)
	}
	deleteFile(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// no such file or directory
		err = err
		return
	}
	t.Error("Delete not sucessfully, ", err)
}

func TestDecodeByStr(t *testing.T) {
	// Having deleted from CS
	empty := `{"key":"","value":""}`
	cnf, err := decodeByStr(empty)
	if err != nil {
		t.Errorf("Having deleted from CS, decoding had one error.")
	} else {
		assertEmptyCS(t, cnf)
	}
}

func TestDecodeSimple(t *testing.T) {
	json := `
          {"key":"com.mogujie.tesla.proxy_MogujieThread1_dynamicConfig","md5":"0bae7dd77b681aa2bc8c49d8e83a5865","value":"{\"failover\":1,\"methodTimeout\":null,\"timeout\":5000}"}  
    `
	expectedSimple := &SimpleConfiguration{
		Key:      "com.mogujie.tesla.proxy_MogujieThread1_dynamicConfig",
		Md5Human: "0bae7dd77b681aa2bc8c49d8e83a5865",
		Value:    "{\"failover\":1,\"methodTimeout\":null,\"timeout\":5000}",
	}
	cfg, err := decodeByStr(json)
	if err != nil {
		t.Errorf("%s", err)
	}
	if !reflect.DeepEqual(cfg, expectedSimple) {
		t.Errorf("Not equal. expected:%q, actual:%q\n", expectedSimple, cfg)
	}
}

func TestDecodeCombined(t *testing.T) {
	// use simply. will be used to put into CombinedConfiguration
	expectedSimple := &SimpleConfiguration{
		Key:      "metaData",
		Md5Human: "0bae7dd77b681aa2bc8c49d8e83a5865",
		Value:    "{\"failover\":1}",
	}
	// Combined, value : []Simple
	json := `
	        {
	            "key": "com.mogujie.tesla.proxy",
	            "md5": "cd8c6eef10d9b3e2d47e9984e41b16d6",
	            "value": [
	                {
	                    "key":"metaData",
	                    "md5":"0bae7dd77b681aa2bc8c49d8e83a5865",
	                    "value": "{\"failover\":1}"
	                }
	            ]
	        }
	    `
	val := []interface{}{expectedSimple}
	expectedCombined := &CombinedConfiguration{
		Key:      "com.mogujie.tesla.proxy",
		Md5Human: "cd8c6eef10d9b3e2d47e9984e41b16d6",
		Value:    val,
	}
	actualComb, err := decodeByStr(json)
	if err != nil {
		t.Errorf("%s", err)
	}
	if !reflect.DeepEqual(expectedCombined, actualComb) {
		t.Errorf("Not equal. expected:%q, actual:%q\n", expectedCombined, actualComb)
	}

	// Combined, value : []Combined
	service1 := &SimpleConfiguration{
		Key:      "RateDsrService_metaData",
		Md5Human: "ac5441d174b81431d9e6d9e9c44aadc3",
		// Value: `{"appName":"rate-service","providerList":[{"ip":"192.168.5.131","port":20020,"version":"1.0.0","weight":100}]}`,
		Value: "appname:rate-service1",
	}
	service2 := &SimpleConfiguration{
		Key:      "RateReadService_dynamicConfig",
		Md5Human: "468f6aef722f0faad55ab19fc07f4d9a",
		// Value: `{"failover":1,"methodTimeout":null,"timeout":1000}`,
		Value: "appname:rate-service2",
	}
	srvV := []interface{}{service1, service2}
	sub1Comb := &CombinedConfiguration{
		Key:      "com.mogujie.tesla.proxy_com.mogujie.service.rate.api.RateDsrService",
		Md5Human: "2e383602ee325476c42350d6d10c65d9",
		Value:    srvV,
	}

	val = make([]interface{}, 0)
	val = append(val, sub1Comb)
	topComb1 := &CombinedConfiguration{
		Key:      "com.mogujie.tesla.proxy",
		Md5Human: "ef8c6eef10d9b3e2d47e9984e41b16c7",
		Value:    val,
	}
	json = `
	        {
	            "key": "com.mogujie.tesla.proxy",
	            "md5": "ef8c6eef10d9b3e2d47e9984e41b16c7",
	            "value": [
	                {
	                    "key":"com.mogujie.tesla.proxy_com.mogujie.service.rate.api.RateDsrService",
	                    "md5":"2e383602ee325476c42350d6d10c65d9",
                        "value":[
                            {
                                "key": "RateDsrService_metaData",
                                "md5": "ac5441d174b81431d9e6d9e9c44aadc3",
                                "value": "appname:rate-service1"
                            },{
                                "key": "RateReadService_dynamicConfig",
                                "md5": "468f6aef722f0faad55ab19fc07f4d9a",
                                "value": "appname:rate-service2"
                            }
                        ]}
	            ]
	        }
    `
	actualTopComb1, err := decodeByStr(json)
	if err != nil {
		t.Errorf("%s", err)
	}
	if !reflect.DeepEqual(actualTopComb1, topComb1) {
		t.Errorf("Not equal. expected:%q, actual:%q\n", topComb1, actualTopComb1)
	}

	// Combined, value : [ Simple , Combined ]
	json = `
	        {
	            "key": "com.mogujie.tesla.proxy",
	            "md5": "ef8c6eef10d9b3e2d47e9984e41b16c7",
	            "value": [
	                {
	                    "key":"com.mogujie.tesla.proxy_com.mogujie.service.rate.api.RateDsrService",
	                    "md5":"2e383602ee325476c42350d6d10c65d9",
                        "value":[
                            {
                                "key": "RateDsrService_metaData",
                                "md5": "ac5441d174b81431d9e6d9e9c44aadc3",
                                "value": "appname:rate-service1"
                            },{
                                "key": "RateReadService_dynamicConfig",
                                "md5": "468f6aef722f0faad55ab19fc07f4d9a",
                                "value": "appname:rate-service2"
                            }
                        ]
                    },
                    {
                        "key": "BeautyService3",
                        "md5": "bc5e41d174b81431d9e6d9e9c44eadcf",
                        "value": "appname:beauty-service3"
                    }
	            ]
	        }
    `
	actualTopComb2, err := decodeByStr(json)
	if err != nil {
		t.Errorf("%s", err)
	}
	simple3 := &SimpleConfiguration{
		Key:      "BeautyService3",
		Md5Human: "bc5e41d174b81431d9e6d9e9c44eadcf",
		Value:    "appname:beauty-service3",
	}
	val = make([]interface{}, 0)
	val = append(val, sub1Comb)
	val = append(val, simple3)
	topComb2 := &CombinedConfiguration{
		Key:      "com.mogujie.tesla.proxy",
		Md5Human: "ef8c6eef10d9b3e2d47e9984e41b16c7",
		Value:    val,
	}
	if !reflect.DeepEqual(actualTopComb2, topComb2) {
		t.Errorf("Not equal. expected:%q, actual:%q\n", topComb2, actualTopComb2)
	}
	// Combined, value : if parse [](Simple||Combined),then error
	var comb *CombinedConfiguration
	//assertConfigurationValue(actualComb.GetValue(), t)
	comb = actualComb.(*CombinedConfiguration)
	if concreteVal, ok := comb.GetValue().([]interface{}); ok {
		for _, v := range concreteVal {
			switch vv := v.(type) {
			case *SimpleConfiguration, *CombinedConfiguration:
				{
					// Perfect
				}
			default:
				{
					t.Error(v, vv)
				}
			}
		}
	} else {
		t.Error("Not right.", comb)
	}
	//assertConfigurationValue(actualTopComb1.GetValue(), t)
	comb = actualTopComb1.(*CombinedConfiguration)
	if concreteVal, ok := comb.GetValue().([]interface{}); ok {
		for _, v := range concreteVal {
			switch vv := v.(type) {
			case *SimpleConfiguration, *CombinedConfiguration:
				{
					// Perfect
				}
			default:
				{
					t.Error(v, vv)
				}
			}
		}
	} else {
		t.Error("Not right.", comb)
	}
	//assertConfigurationValue(actualTopComb2.GetValue(), t)
	comb = actualTopComb2.(*CombinedConfiguration)
	if concreteVal, ok := comb.GetValue().([]interface{}); ok {
		for _, v := range concreteVal {
			switch vv := v.(type) {
			case *SimpleConfiguration, *CombinedConfiguration:
				{
					// Perfect
				}
			default:
				{
					t.Error(v, vv)
				}
			}
		}
	} else {
		t.Error("Not right.", comb)
	}
}

func assertEmptyCS(t *testing.T, cnf Configuration) {
	v, ok := cnf.(*SimpleConfiguration)
	if !ok {
		t.Errorf("%q", v)
	} else {
		if cnf.GetKey() != "" {
			t.Errorf("Not empty value.")
		}
		if cnf.GetValue().(string) != "" {
			t.Errorf("Not empty key.")
		}
	}
}
