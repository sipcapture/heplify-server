package config

var Setting HeplifyServer

type HeplifyServer struct {
	Network        string `default:"udp"`
	Protobuf       bool   `default:"false"`
	Mode           string `default:"homer5"`
	HEPAddr        string `default:"0.0.0.0:9060"`
	HEPWorkers     int    `default:"100"`
	MQName         string `default:""`
	MQAddr         string `default:""`
	MQTopic        string `default:""`
	PromAddr       string `default:""`
	PromTargetIP   string `default:""`
	PromTargetName string `default:""`
	DBDriver       string `default:"mysql"`
	DBAddr         string `default:"localhost:3306"`
	DBUser         string `default:"root"`
	DBPass         string `default:""`
	DBDataTable    string `default:"homer_data"`
	DBConfTable    string `default:"homer_configuration"`
	DBPath         string `default:""`
	DBBulk         int    `default:"200"`
	DBTimer        int    `default:"2"`
	DBRotate       bool   `default:"true"`
	DBRotateLog    string `default:"6h"`
	DBRotateSip    string `default:"2h"`
	DBRotateQos    string `default:"12h"`
	DBDropDays     int    `default:"0"`
	Dedup          bool   `default:"false"`
	SentryDSN      string `default:""`
	AlegID         string `default:"x-cid"`
	LogDbg         string `default:""`
	LogLvl         string `default:"info"`
	LogStd         bool   `default:"false"`
	Config         string `default:"./heplify-server.toml"`
	Version        bool   `default:"false"`
}

func NewConfig() *HeplifyServer {
	return &HeplifyServer{
		Network:        "udp",
		Protobuf:       false,
		Mode:           "homer5",
		HEPAddr:        "0.0.0.0:9060",
		HEPWorkers:     100,
		MQName:         "",
		MQAddr:         "",
		MQTopic:        "",
		PromAddr:       "",
		PromTargetIP:   "",
		PromTargetName: "",
		DBDriver:       "mysql",
		DBAddr:         "localhost:3306",
		DBUser:         "root",
		DBPass:         "",
		DBDataTable:    "homer_data",
		DBConfTable:    "homer_configuration",
		DBPath:         "",
		DBBulk:         200,
		DBTimer:        2,
		DBRotate:       true,
		DBRotateLog:    "6h",
		DBRotateSip:    "2h",
		DBRotateQos:    "12h",
		DBDropDays:     0,
		Dedup:          false,
		SentryDSN:      "",
		AlegID:         "x-cid",
		LogDbg:         "",
		LogLvl:         "info",
		LogStd:         false,
		Config:         "./heplify-server.toml",
		Version:        false,
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
