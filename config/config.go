package config

var Setting HeplifyServer

type HeplifyServer struct {
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
	DBBulk         int    `default:"100"`
	DBRotate       bool   `default:"true"`
	DBDropDays     int    `default:"0"`
	SentryDSN      string `default:""`
	AlegID         string `default:"x-cid"`
	LogDbg         string `default:""`
	LogLvl         string `default:"info"`
	LogStd         bool   `default:"false"`
	Config         string `default:"./heplify-server.toml"`
}

func NewConfig() *HeplifyServer {
	return &HeplifyServer{
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
		DBBulk:         100,
		DBRotate:       true,
		DBDropDays:     0,
		SentryDSN:      "",
		AlegID:         "x-cid",
		LogDbg:         "",
		LogLvl:         "info",
		LogStd:         false,
		Config:         "./heplify-server.toml",
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
