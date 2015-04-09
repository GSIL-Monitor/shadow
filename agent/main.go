package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
	// third pkgs
	log "github.com/cihub/seelog"
	// company pkgs
	cnf "shadow/agent/cnf"
	tran "shadow/agent/transshipment"
)

var (
	listener net.Listener
)

func init() {
	log.Info("Initialize: main")
	rand.Seed(time.Now().UnixNano())
}

func main() {
	defer log.Flush()
	flag.Parse()

	log.Info("Application is booting...")
	var err error
	// seelog
	f := filepath.Join(cnf.CnfBasedir, "seelog.xml")
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		log.Error(err)
	} else {
		if logger, err := log.LoggerFromConfigAsBytes(bs); err == nil {
			log.ReplaceLogger(logger)
		} else {
			log.Error(err)
		}
	}

	// begin business
	// Goroutines > 1 , do not remove > 1
	num := runtime.NumCPU() * 4
	if num == 1 {
		num = 2
	}
	runtime.GOMAXPROCS(num)
	// App some signal handles.
	go guard()
	check()

	log.Info("Application booted. Now listen the socket.")
	dir := "/var/run/shadow/"
	os.MkdirAll(dir, 0755)
listen:
	unix_domain := "/var/run/shadow/agent.sock"
	listener, err = net.Listen("unix", unix_domain)
	if err != nil {
		log.Error(err)
		// Unix sockets must be unlink()ed before being reused again.
		syscall.Unlink(unix_domain)
		time.Sleep(500 * time.Millisecond)
		goto listen
	}
	for {
		// block the main process
		conn, err := listener.Accept()
		if err != nil {
			log.Error("Accept error:", err)
			continue
		}
		go handle(conn)
	}
	// end listen
}

func handle(c net.Conn) {
	tran.Handle(c)
}

// blocking
func guard() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSTOP)
	for {
		s := <-ch
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGSTOP, syscall.SIGINT:
			log.Info("Application got signals.")
			if listener != nil {
				listener.Close()
			}
			os.Exit(0)
			return
		case syscall.SIGHUP:
			reload()
			return
		default:
			return
		}
	}
}

//TODO
func reload() {
}

//TODO
func check() {
}
