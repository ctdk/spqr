package config

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/tideland/golib/logger"
	"log"
	"os"
	"runtime"
)

const Version = "0.0.1"

const defaultUserKeyPrefix = "org/default/users"

var debugLevelDesc = map[int]string{0: "debug", 1: "info", 2: "warning", 3: "error", 4: "critical", 5: "fatal"}

var GitHash = "unknown"

type Conf struct {
	ConsulHttpAddr string
	UserKeyPrefix string
	DebugLevel int
	LogFile string
	SysLog bool
	StateFile string
}

type Options struct {
	Version bool `short:"v" long:"version" description:"Print version info."`
	ConsulHttpAddr string `short:"C" long:"consul-http-addr" description:"Consul HTTP API endpoint. Defaults to http://127.0.0.1:8500. Shares the same CONSUL_HTTP_ADDR environment variable as consul itself as well." env:"CONSUL_HTTP_ADDR"`
	UserKeyPrefix string `short:"P" long:"user-key-prefix" description:"Consul key prefix for user data. Default value: 'org/default/users'." env:"SPQR_USER_KEY_PREFIX"`
	LogFile string `short:"L" long:"log-file" description:"Log to file X" env:"SPQR_LOG_FILE"`
	SysLog bool `short:"S" long:"syslog" description:"Log to syslog rather than to a log file. Incompatible with -L/--log-file." env:"SPQR_SYSLOG"`
	StateFile string `short:"s" long:"statefile" description:"Store spqr's state in this file."`
	Verbose              []bool       `short:"V" long:"verbose" description:"Show verbose debug information. Repeat for more verbosity."`
}

func initConfig() *Conf { return &Conf{} }

var Config = initConfig()

func ParseConfigOptions() error {
	var opts = &Options{}
	parser := flags.NewParser(opts, flags.Default)
	parser.ShortDescription = fmt.Sprintf("A consul leveraging user account manager - version %s", Version)

	parser.NamespaceDelimiter = "-"

	_, err := parser.Parse()
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			log.Println(err)
			os.Exit(1)
		}
	}
	if opts.Version {
		fmt.Printf("spqr version %s (git hash: %s) built with %s.\n", Version, GitHash, runtime.Version())
		os.Exit(0)
	}
	if opts.ConsulHttpAddr != "" {
		Config.ConsulHttpAddr = opts.ConsulHttpAddr
	}
	if Config.ConsulHttpAddr == "" {
		Config.ConsulHttpAddr = "http://127.0.0.1:8500"
	}

	if opts.UserKeyPrefix != "" {
		Config.UserKeyPrefix = opts.UserKeyPrefix
	}
	if Config.UserKeyPrefix == "" {
		Config.UserKeyPrefix = defaultUserKeyPrefix
	}

	if opts.SysLog {
		Config.SysLog = opts.SysLog
	}
	if Config.LogFile != "" {
		lfp, lerr := os.OpenFile(Config.LogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModeAppend|0666)
		if lerr != nil {
			log.Println(err)
			os.Exit(1)
		}
		log.SetOutput(lfp)
	}

	if dl := len(opts.Verbose); dl != 0 {
		Config.DebugLevel = dl
	}
	if Config.DebugLevel > 5 {
		Config.DebugLevel = 5
	}
	Config.DebugLevel = int(logger.LevelFatal) - Config.DebugLevel
	logger.SetLevel(logger.LogLevel(Config.DebugLevel))
	log.Printf("Logging at %s level", debugLevelDesc[Config.DebugLevel])
	lerr := setLogger(Config.SysLog)
	if lerr != nil {
		log.Println(lerr.Error())
		os.Exit(1)
	}

	if opts.StateFile != "" {
		Config.StateFile = opts.StateFile
	}

	return nil
}
