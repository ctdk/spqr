// +build windows

package config

import (
	"github.com/tideland/golib/logger"
	"log"
)

func setLogger(useSyslog bool) error {
	if useSyslog {
		log.Println("Syslog isn't actually supported in Windows - using regular logging.")
	}
	logger.SetLogger(logger.NewGoLogger())
	return nil
}
