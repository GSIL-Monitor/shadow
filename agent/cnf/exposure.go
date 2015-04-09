package cnf

import (
	"flag"

	// third pkgs
	log "github.com/cihub/seelog"
)

var (
	CnfBasedir string
)

func init() {
	flag.StringVar(&CnfBasedir, "cnf_basedir", "/", "Please input the absolute path of application configuration dir.")
	if CnfBasedir == "" {
		log.Error("Not expected CnfBasedir:", CnfBasedir)
	}
}
