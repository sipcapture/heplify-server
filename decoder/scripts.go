package decoder

import (
	"fmt"
	"io/ioutil"

	"github.com/negbie/logp"
	"github.com/sipcapture/golua/lua"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder/luar"
)

/// structure for Script Engine
type ScriptEngine struct {
	script    string
	LuaEngine *lua.State
	mapObj    luar.Map
	/* pointer to modify */
	hepPkt **HEP
}

func (d *ScriptEngine) getSIPObject() interface{} {
	return (*d.hepPkt).SIP
}

func (d *ScriptEngine) getHEPProtoType() uint32 {
	return (*d.hepPkt).GetProtoType()
}

func (d *ScriptEngine) getHEPObject() interface{} {
	return (*d.hepPkt)
}

func (d *ScriptEngine) getHEPSrcIP() string {
	return (*d.hepPkt).GetSrcIP()
}

func (d *ScriptEngine) getHEPSrcPort() uint32 {
	return (*d.hepPkt).GetSrcPort()
}

func (d *ScriptEngine) getHEPDstIP() string {
	return (*d.hepPkt).GetDstIP()
}

func (d *ScriptEngine) getHEPDstPort() uint32 {
	return (*d.hepPkt).GetDstPort()
}

func (d *ScriptEngine) getHEPTimeSeconds() uint32 {
	return (*d.hepPkt).GetTsec()
}

func (d *ScriptEngine) getHEPTimeUseconds() uint32 {
	return (*d.hepPkt).GetTmsec()
}

func (d *ScriptEngine) setCustomHeaders(m *map[string]string) {
	hepPkt := *d.hepPkt

	/* not SIP */
	if hepPkt.SIP == nil {
		return
	}

	if hepPkt.SIP.CustomHeader == nil {
		hepPkt.SIP.CustomHeader = make(map[string]string)
	}

	for k, v := range *m {
		hepPkt.SIP.CustomHeader[k] = v
	}
}

func (d *ScriptEngine) getRawMessage() string {
	hepPkt := *d.hepPkt
	return hepPkt.Payload
}

func (d *ScriptEngine) applyHeader(header string, value string) {
	hepPkt := *d.hepPkt

	switch {
	case header == "Via":
		hepPkt.SIP.ViaOne = value
	case header == "FromUser":
		hepPkt.SIP.FromUser = value
	case header == "FromHost":
		hepPkt.SIP.FromHost = value
	case header == "FromTag":
		hepPkt.SIP.FromTag = value
	case header == "ToUser":
		hepPkt.SIP.ToUser = value
	case header == "ToHost":
		hepPkt.SIP.ToHost = value
	case header == "ToTag":
		hepPkt.SIP.ToTag = value
	case header == "Call-ID":
		hepPkt.SIP.CallID = value
	case header == "X-CID":
		hepPkt.SIP.XCallID = value
	case header == "ContactUser":
		hepPkt.SIP.ContactUser = value
	case header == "ContactHost":
		hepPkt.SIP.ContactHost = value
	case header == "User-Agent":
		hepPkt.SIP.UserAgent = value
	case header == "Server":
		hepPkt.SIP.Server = value
	case header == "AuthorizationUsername":
		hepPkt.SIP.Authorization.Username = value
	case header == "Proxy-AuthorizationUsername":
		hepPkt.SIP.Authorization.Username = value
	case header == "PAIUser":
		hepPkt.SIP.PaiUser = value
	case header == "PAIHost":
		hepPkt.SIP.PaiHost = value
	case header == "RAW":
		hepPkt.Payload = value
	}
}

func (d *ScriptEngine) logData(level string, message string, data interface{}) {
	if level == "ERROR" {
		logp.Err("[script] %s: %v", message, data)
	} else {
		logp.Debug("[script] %s: %v", message, data)
	}
}

// RegisteredScriptEngine returns a script interface
func RegisteredScriptEngine() (*ScriptEngine, error) {
	logp.Debug("script", "Init script engine...")

	dec := &ScriptEngine{}
	dec.LuaEngine = lua.NewState()
	dec.LuaEngine.OpenLibs()

	luar.Register(dec.LuaEngine, "HEP", luar.Map{
		"applyHeader":        dec.applyHeader,
		"setCustomHeaders":   dec.setCustomHeaders,
		"getSIPObject":       dec.getSIPObject,
		"getHEPProtoType":    dec.getHEPProtoType,
		"getHEPSrcIP":        dec.getHEPSrcIP,
		"getHEPSrcPort":      dec.getHEPSrcPort,
		"getHEPDstIP":        dec.getHEPDstIP,
		"getHEPDstPort":      dec.getHEPDstPort,
		"getHEPTimeSeconds":  dec.getHEPTimeSeconds,
		"getHEPTimeUseconds": dec.getHEPTimeUseconds,
		"getHEPObject":       dec.getHEPObject,
		"getRawMessage":      dec.getRawMessage,
		"logData":            dec.logData,
		"print":              fmt.Println,
	})

	data, err := ioutil.ReadFile(config.Setting.ScriptFile)
	if err != nil {
		return nil, err
	}

	dec.script = string(data)
	dec.LuaEngine.DoString(dec.script)
	return dec, nil
}

/* our main point ExecuteScriptEngine */
func (d *ScriptEngine) ExecuteScriptEngine(hep *HEP) error {
	/* preload */
	d.hepPkt = &hep
	return d.LuaEngine.DoString("init()")
}
