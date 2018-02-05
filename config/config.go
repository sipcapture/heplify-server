package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/negbie/heplify/logp"
)

var (
	Cfg         Config
	logging     logp.Logging
	fileRotator logp.FileRotator
)

const version = "heplify-server 0.01"

type Config struct {
	HEPPort       int
	HEPWorkers    int
	HEPTopic      string
	OutName       string
	OutConfigFile string
}

func NewConfig() *Config {
	return &Config{
		HEPPort:       9060,
		HEPWorkers:    100,
		HEPTopic:      "HEP",
		OutName:       "nsq",
		OutConfigFile: "/etc/heplify-server/output.conf",
	}
}

func Get() *Config {
	return NewConfig()
}

func (conf *Config) ParseFlags() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Use %s like: %s [option]\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	logging.Files = &fileRotator

	flag.StringVar(&logging.Level, "l", "info", "Log level [debug, info, warning, error]")
	flag.StringVar(&fileRotator.Path, "p", "./", "Log filepath")
	flag.StringVar(&fileRotator.Name, "n", "heplify-server.log", "Log filename")
	flag.IntVar(&conf.HEPPort, "hp", conf.HEPPort, "HEP port")
	flag.IntVar(&conf.HEPWorkers, "hw", conf.HEPWorkers, "HEP workers")
	flag.StringVar(&conf.HEPTopic, "ht", conf.HEPTopic, "HEP topic name")
	flag.StringVar(&conf.OutName, "qn", conf.OutName, "output name")
	flag.StringVar(&conf.OutConfigFile, "qc", conf.OutConfigFile, "output config file")

	flag.Parse()
}
