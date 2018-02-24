package config

import "github.com/negbie/heplify-server/logp"

var Setting Config

type Config struct {
	HEPAddr    string
	HEPWorkers int

	MQName  string
	MQAddr  string
	MQTopic string

	PromAddr   string
	PromTarget string

	DBDriver   string
	DBAddr     string
	DBBulk     int
	DBName     string
	DBUser     string
	DBPassword string
	Logging    *logp.Logging
}

func NewConfig() *Config {
	return &Config{

		HEPAddr:    "0.0.0.0:9060",
		HEPWorkers: 100,

		MQName:  "nsq",
		MQAddr:  "localhost:4015",
		MQTopic: "hep",

		PromAddr:   "0.0.0.0:8888",
		PromTarget: "",

		DBDriver:   "mysql",
		DBAddr:     "localhost:3306",
		DBBulk:     100,
		DBName:     "homer_data",
		DBUser:     "test",
		DBPassword: "test",
	}
}

func Get() *Config {
	return NewConfig()
}
