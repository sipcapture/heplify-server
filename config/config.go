package config

const Version = "heplify-server 0.952"

var Setting HeplifyServer

type HeplifyServer struct {
	HEPAddr         string   `default:"0.0.0.0:9060"`
	ESAddr          string   `default:""`
	ESDiscovery     bool     `default:"true"`
	MQDriver        string   `default:""`
	MQAddr          string   `default:""`
	MQTopic         string   `default:""`
	PromAddr        string   `default:""`
	PromTargetIP    string   `default:""`
	PromTargetName  string   `default:""`
	HoraclifixStats bool     `default:"false"`
	RTPAgentStats   bool     `default:"false"`
	DBShema         string   `default:"homer5"`
	DBDriver        string   `default:"mysql"`
	DBAddr          string   `default:"localhost:3306"`
	DBUser          string   `default:"root"`
	DBPass          string   `default:""`
	DBDataTable     string   `default:"homer_data"`
	DBConfTable     string   `default:"homer_configuration"`
	DBTableSpace    string   `default:""`
	DBBulk          int      `default:"200"`
	DBTimer         int      `default:"2"`
	DBRotate        bool     `default:"true"`
	DBPartLog       string   `default:"2h"`
	DBPartSip       string   `default:"1h"`
	DBPartQos       string   `default:"6h"`
	DBDropDays      int      `default:"0"`
	DBDropOnStart   bool     `default:"false"`
	Dedup           bool     `default:"false"`
	DiscardMethod   []string `default:""`
	AlegIDs         []string `default:""`
	LogDbg          string   `default:""`
	LogLvl          string   `default:"info"`
	LogStd          bool     `default:"false"`
	Config          string   `default:"./heplify-server.toml"`
	Version         bool     `default:"false"`
}

func NewConfig() *HeplifyServer {
	return &HeplifyServer{
		HEPAddr:         "0.0.0.0:9060",
		ESAddr:          "",
		ESDiscovery:     true,
		MQDriver:        "",
		MQAddr:          "",
		MQTopic:         "",
		PromAddr:        "",
		PromTargetIP:    "",
		PromTargetName:  "",
		HoraclifixStats: false,
		RTPAgentStats:   false,
		DBShema:         "homer5",
		DBDriver:        "mysql",
		DBAddr:          "localhost:3306",
		DBUser:          "root",
		DBPass:          "",
		DBDataTable:     "homer_data",
		DBConfTable:     "homer_configuration",
		DBTableSpace:    "",
		DBBulk:          200,
		DBTimer:         2,
		DBRotate:        true,
		DBPartLog:       "6h",
		DBPartSip:       "2h",
		DBPartQos:       "12h",
		DBDropDays:      0,
		DBDropOnStart:   false,
		Dedup:           false,
		DiscardMethod:   nil,
		AlegIDs:         nil,
		LogDbg:          "",
		LogLvl:          "info",
		LogStd:          false,
		Config:          "./heplify-server.toml",
		Version:         false,
	}
}

func Get() *HeplifyServer {
	return NewConfig()
}
