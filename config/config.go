package config

var Setting HeplifyServer

type HeplifyServer struct {
	HEPAddr    string `default:"0.0.0.0:9060"`
	HEPWorkers int    `default:"100"`

	MQName  string `default:""`
	MQAddr  string `default:""`
	MQTopic string `default:""`

	PromAddr       string `default:""`
	PromHunterIP   string `default:""`
	PromHunterName string `default:""`

	DBDriver string `default:"mysql"`
	DBAddr   string `default:"localhost:3306"`
	DBBulk   int    `default:"100"`
	DBName   string `default:"homer_data"`
	DBUser   string `default:"test"`
	DBPass   string `default:"test"`

	AlegID string `default:"x-cid"`

	Loglvl string `default:"info"`
	LogStd bool   `default:"false"`
	Logdbg string `default:""`
}

func NewConfig() *HeplifyServer {
	return &HeplifyServer{

		HEPAddr:    "0.0.0.0:9060",
		HEPWorkers: 100,

		MQName:  "",
		MQAddr:  "",
		MQTopic: "",

		PromAddr:       "",
		PromHunterIP:   "",
		PromHunterName: "",

		DBDriver: "mysql",
		DBAddr:   "localhost:3306",
		DBBulk:   100,
		DBName:   "homer_data",
		DBUser:   "test",
		DBPass:   "test",

		AlegID: "x-cid",

		Loglvl: "info",
		LogStd: false,
		Logdbg: "",
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
