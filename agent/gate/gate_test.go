package gate

// Integration Testing
// Continuous Integration
import (
	"flag"
	"log"
	"os"
	"testing"
)

var (
	env = os.Getenv("APP_ENV")
)

func init() {
	flag.Parse()
	log.Println("APP_ENV: ", env)
}

// func getApps() []string
func TestGetApps(t *testing.T) {
	apps := getApps()
	switch env {
	case "dev", "develop":
		{
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
	case "test":
		{
		}
	case "pro", "production":
		{
		}
	default:
		{
			t.Error("Not expected APP_ENV:", env)
		}
	}
}

// func initListeners()
func TestInitListeners(t *testing.T) {
	if env == "test" {
	}
}

// func doCheckRings()
func TestDoCheckRings(t *testing.T) {
	if env == "test" {
	}
}

// func (gate *Gate) Close() error
func TestGateClose(t *testing.T) {
	if env == "test" {
	}
}

// func getRingName(app string) (string, error)
func TestGetRingName(t *testing.T) {
	if env == "test" {
	}
}

// func getInstancesFromCS(ring string) map[string]uint32
func TestGetInstancesFromCS(t *testing.T) {
	if env == "test" {
	}
}

// func getInstances(ringName, yaml string) map[string]uint32
// From CS
func TesetGetInstances(t *testing.T) {
	if env == "test" {
	}
}

// func newPool(server, password string) *redis.Pool
func TestNewPool(t *testing.T) {
	if env == "test" {
	}
}

// func newPools(instances map[string]uint32) (pools map[string]*redis.Pool, err error)
func TestNewPools(t *testing.T) {
	if env == "test" {
	}
}

// func (r *ring) Destroy()
func TestRingDestroy(t *testing.T) {
	if env == "test" {
	}
}

// func newRing(instances map[string]uint32) (r *ring, err error)
func TestNewRing(t *testing.T) {
	if env == "test" {
	}
}
