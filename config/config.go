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
	DBData         string `default:"homer_data"`
	DBConf         string `default:"homer_configuration"`
	DBUser         string `default:"root"`
	DBPass         string `default:""`
	DBPath         string `default:""`
	DBBulk         int    `default:"100"`
	SentryDSN      string `default:""`
	AlegID         string `default:"x-cid"`
	LogDbg         string `default:""`
	LogLvl         string `default:"info"`
	LogStd         bool   `default:"false"`
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
		DBData:         "homer_data",
		DBConf:         "homer_configuration",
		DBUser:         "root",
		DBPass:         "",
		DBPath:         "",
		DBBulk:         100,
		SentryDSN:      "",
		AlegID:         "x-cid",
		LogDbg:         "",
		LogLvl:         "info",
		LogStd:         false,
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
