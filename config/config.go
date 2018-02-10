package config

import "github.com/negbie/heplify-server/logp"

var Setting Config

type Config struct {
	HEPAddr    string
	HEPWorkers int
	HEPTopic   string

	NSQName string
	NSQAddr string

	DBAddr     string
	DBName     string
	DBUser     string
	DBPassword string
	Logging    *logp.Logging
}

func NewConfig() *Config {
	return &Config{
		HEPAddr:    "localhost:9060",
		HEPWorkers: 100,
		HEPTopic:   "hep",

		NSQName: "nsq",
		NSQAddr: "localhost:4015",

		DBAddr:     "localhost:3306",
		DBName:     "homer_data",
		DBUser:     "test",
		DBPassword: "test",
	}
}

func Get() *Config {
	return NewConfig()
}
