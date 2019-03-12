package config

const Version = "heplify-server 1.07"

var Setting HeplifyServer

type HeplifyServer struct {
	HEPAddr            string   `default:"0.0.0.0:9060"`
	HEPTCPAddr         string   `default:""`
	HEPTLSAddr         string   `default:"0.0.0.0:9060"`
	HTTPAddr           string   `default:""`
	ESAddr             string   `default:""`
	ESDiscovery        bool     `default:"true"`
	MQDriver           string   `default:""`
	MQAddr             string   `default:""`
	MQTopic            string   `default:""`
	LokiURL            string   `default:""`
	LokiBulk           int      `default:"200"`
	LokiTimer          int      `default:"2"`
	LokiBuffer         int      `default:"100000"`
	LokiHEPFilter      []int    `default:"1,100"`
	PromAddr           string   `default:""`
	PromTargetIP       string   `default:""`
	PromTargetName     string   `default:""`
	DBShema            string   `default:"homer5"`
	DBDriver           string   `default:"mysql"`
	DBAddr             string   `default:"localhost:3306"`
	DBUser             string   `default:"root"`
	DBPass             string   `default:""`
	DBDataTable        string   `default:"homer_data"`
	DBConfTable        string   `default:"homer_configuration"`
	DBTableSpace       string   `default:""`
	DBBulk             int      `default:"400"`
	DBTimer            int      `default:"4"`
	DBBuffer           int      `default:"400000"`
	DBWorker           int      `default:"8"`
	DBRotate           bool     `default:"true"`
	DBPartLog          string   `default:"2h"`
	DBPartIsup         string   `default:"6h"`
	DBPartSip          string   `default:"2h"`
	DBPartQos          string   `default:"6h"`
	DBDropDays         int      `default:"14"`
	DBDropDaysCall     int      `default:"0"`
	DBDropDaysRegister int      `default:"0"`
	DBDropDaysDefault  int      `default:"0"`
	DBDropOnStart      bool     `default:"false"`
	Dedup              bool     `default:"false"`
	DiscardMethod      []string `default:""`
	FilterHost         []string `default:""`
	AlegIDs            []string `default:""`
	LogDbg             string   `default:""`
	LogLvl             string   `default:"info"`
	LogStd             bool     `default:"false"`
	Config             string   `default:"./heplify-server.toml"`
	ConfigHTTPAddr     string   `default:""`
}

func NewConfig() *HeplifyServer {
	return &HeplifyServer{
		HEPAddr:            "0.0.0.0:9060",
		HEPTCPAddr:         "",
		HEPTLSAddr:         "0.0.0.0:9060",
		HTTPAddr:           "",
		ESAddr:             "",
		ESDiscovery:        true,
		MQDriver:           "",
		MQAddr:             "",
		MQTopic:            "",
		LokiURL:            "",
		LokiBulk:           200,
		LokiTimer:          2,
		PromAddr:           "",
		PromTargetIP:       "",
		PromTargetName:     "",
		DBShema:            "homer5",
		DBDriver:           "mysql",
		DBAddr:             "localhost:3306",
		DBUser:             "root",
		DBPass:             "",
		DBDataTable:        "homer_data",
		DBConfTable:        "homer_configuration",
		DBTableSpace:       "",
		DBBulk:             200,
		DBTimer:            2,
		DBBuffer:           400000,
		DBWorker:           8,
		DBRotate:           true,
		DBPartLog:          "6h",
		DBPartIsup:         "6h",
		DBPartSip:          "2h",
		DBPartQos:          "12h",
		DBDropDays:         14,
		DBDropDaysCall:     0,
		DBDropDaysRegister: 0,
		DBDropDaysDefault:  0,
		DBDropOnStart:      false,
		Dedup:              false,
		DiscardMethod:      nil,
		FilterHost:         nil,
		AlegIDs:            nil,
		LogDbg:             "",
		LogLvl:             "info",
		LogStd:             false,
		Config:             "./heplify-server.toml",
		ConfigHTTPAddr:     "",
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
