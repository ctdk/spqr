// +build !windows

package config

import (
	"github.com/tideland/golib/logger"
)

func setLogger(useSyslog bool) error {
	if useSyslog {
		sl, err := logger.NewSysLogger("spqr")
		if err != nil {
			return err
		}
		logger.SetLogger(sl)
	} else {
		logger.SetLogger(logger.NewGoLogger())
	}
	return nil
}
