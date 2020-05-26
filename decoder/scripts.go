package decoder

import (
	"fmt"
	"reflect"

	"github.com/aarzilli/golua/lua"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/stevedonovan/luar"
)

/// structure for Script Engine
type ScriptEngine struct {
	LuaEngine *lua.State
	mapObj    luar.Map
	/* pointer to modify */
	hepPkt **HEP
}

func (d *ScriptEngine) getParsedVariables() interface{} {

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

	return
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

	return
}

func (d *ScriptEngine) logData(level string, message string, data interface{}) {

	if level == "ERROR" {
		logp.Err("script - log", "%s: %v", message, data)
	} else {
		logp.Debug("script - log", "%s: %v", message, data)
	}
}

// RegisteredScriptEngine returns a script interface
func RegisteredScriptEngine() (*ScriptEngine, error) {

	dec := &ScriptEngine{}

	dec.LuaEngine = luar.Init()

	API := luar.Map{
		"applyHeader":        dec.applyHeader,
		"logData":            dec.logData,
		"setCustomHeaders":   dec.setCustomHeaders,
		"getParsedVariables": dec.getParsedVariables,
		"getHEPProtoType":    dec.getHEPProtoType,
		"getHEPSrcIP":        dec.getHEPSrcIP,
		"getHEPSrcPort":      dec.getHEPSrcPort,
		"getHEPDstIP":        dec.getHEPDstIP,
		"getHEPDstPort":      dec.getHEPDstPort,
		"getHEPTimeSeconds":  dec.getHEPTimeSeconds,
		"getHEPTimeUseconds": dec.getHEPTimeUseconds,
		"getHEPObject":       dec.getHEPObject,
		"getRawMessage":      dec.getRawMessage,
		"print":              fmt.Println,
	}

	dec.mapObj = make(luar.Map)
	dec.mapObj["api"] = API

	logp.Debug("script", "Init script engine...")

	luar.Register(dec.LuaEngine, "HEP", dec.mapObj)

	/* LOAD */
	if r := dec.LuaEngine.LoadFile(config.Setting.ScriptFile); r != 0 {
		logp.Err("script", "ERROR: %v", dec.LuaEngine.ToString(-1))
		return nil, &reflect.ValueError{}
	}

	dec.LuaEngine.Call(0, 0)

	return dec, nil

}

/* our main point ExecuteScriptEngine */
func (d *ScriptEngine) ExecuteScriptEngine(hep *HEP) {

	/* preload */
	d.hepPkt = &hep

	d.LuaEngine.GetGlobal("init")

	logp.Debug("script", "%+v\n\n")

	err := d.LuaEngine.Call(1, 0)
	if err != nil {
		logp.Err("Execute script failed", "%v\n", err)
		return
	}

	logp.Debug("script", "%+v\n\n")
}
