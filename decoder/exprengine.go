package decoder

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/negbie/logp"
)

/// structure for Script Engine
type ExprEngine struct {
	/* pointer to modify */
	hepPkt **HEP
	prog   *vm.Program
	env    map[string]interface{}
	v      vm.VM
}

func (d *ExprEngine) GetHEPStruct() interface{} {
	if (*d.hepPkt) == nil {
		return ""
	}
	return (*d.hepPkt)
}

func (d *ExprEngine) GetSIPStruct() interface{} {
	if (*d.hepPkt).SIP == nil {
		return ""
	}
	return (*d.hepPkt).SIP
}

func (d *ExprEngine) GetHEPProtoType() uint32 {
	return (*d.hepPkt).GetProtoType()
}

func (d *ExprEngine) GetHEPSrcIP() string {
	return (*d.hepPkt).GetSrcIP()
}

func (d *ExprEngine) GetHEPSrcPort() uint32 {
	return (*d.hepPkt).GetSrcPort()
}

func (d *ExprEngine) GetHEPDstIP() string {
	return (*d.hepPkt).GetDstIP()
}

func (d *ExprEngine) GetHEPDstPort() uint32 {
	return (*d.hepPkt).GetDstPort()
}

func (d *ExprEngine) GetHEPTimeSeconds() uint32 {
	return (*d.hepPkt).GetTsec()
}

func (d *ExprEngine) GetHEPTimeUseconds() uint32 {
	return (*d.hepPkt).GetTmsec()
}

func (d *ExprEngine) GetHEPNodeID() uint32 {
	return (*d.hepPkt).GetNodeID()
}

func (d *ExprEngine) GetRawMessage() string {
	return (*d.hepPkt).GetPayload()
}

func (d *ExprEngine) SetRawMessage(value string) error {
	if (*d.hepPkt) == nil {
		err := fmt.Errorf("can't set Raw message if HEP struct is nil, please check for nil in lua script\n")
		logp.Err("%v", err)
		return err
	}
	hepPkt := *d.hepPkt
	hepPkt.Payload = value
	return nil
}

func (d *ExprEngine) SetCustomSIPHeader(m *map[string]string) error {
	if (*d.hepPkt).SIP == nil {
		err := fmt.Errorf("can't set custom SIP header if SIP struct is nil, please check for nil in lua script\n")
		logp.Err("%v", err)
		return err
	}
	hepPkt := *d.hepPkt

	if hepPkt.SIP.CustomHeader == nil {
		hepPkt.SIP.CustomHeader = make(map[string]string)
	}

	for k, v := range *m {
		hepPkt.SIP.CustomHeader[k] = v
	}
	return nil
}

func (d *ExprEngine) SetHEPField(field string, value string) error {
	if (*d.hepPkt) == nil {
		err := fmt.Errorf("can't set HEP field if HEP struct is nil, please check for nil in lua script\n")
		logp.Err("%v", err)
		return err
	}
	hepPkt := *d.hepPkt

	switch field {
	case "ProtoType":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.ProtoType = uint32(i)
		}
	case "SrcIP":
		hepPkt.SrcIP = value
	case "SrcPort":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.SrcPort = uint32(i)
		}
	case "DstIP":
		hepPkt.DstIP = value
	case "DstPort":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.DstPort = uint32(i)
		}
	case "NodeID":
		if i, err := strconv.Atoi(value); err == nil {
			hepPkt.NodeID = uint32(i)
		}
	case "CID":
		hepPkt.CID = value
	case "SID":
		hepPkt.SID = value
	case "NodeName":
		hepPkt.NodeName = value
	case "TargetName":
		hepPkt.TargetName = value
	}
	return nil
}

func (d *ExprEngine) SetSIPHeader(header string, value string) error {
	if (*d.hepPkt).SIP == nil {
		err := fmt.Errorf("can't set SIP header if SIP struct is nil, please check for nil in lua script\n")
		logp.Err("%v", err)
		return err
	}
	hepPkt := *d.hepPkt

	switch header {
	case "ViaOne":
		hepPkt.SIP.ViaOne = value
	case "FromUser":
		hepPkt.SIP.FromUser = value
	case "FromHost":
		hepPkt.SIP.FromHost = value
	case "FromTag":
		hepPkt.SIP.FromTag = value
	case "ToUser":
		hepPkt.SIP.ToUser = value
	case "ToHost":
		hepPkt.SIP.ToHost = value
	case "ToTag":
		hepPkt.SIP.ToTag = value
	case "CallID":
		hepPkt.SIP.CallID = value
	case "XCallID":
		hepPkt.SIP.XCallID = value
	case "ContactUser":
		hepPkt.SIP.ContactUser = value
	case "ContactHost":
		hepPkt.SIP.ContactHost = value
	case "UserAgent":
		hepPkt.SIP.UserAgent = value
	case "Server":
		hepPkt.SIP.Server = value
	case "Authorization.Username":
		hepPkt.SIP.Authorization.Username = value
	case "PaiUser":
		hepPkt.SIP.PaiUser = value
	case "PaiHost":
		hepPkt.SIP.PaiHost = value
	}
	return nil
}

func (d *ExprEngine) Hash(s, name string) string {
	switch name {
	case "md5":
		return fmt.Sprintf("%x", md5.Sum([]byte(s)))
	case "sha1":
		return fmt.Sprintf("%x", sha1.Sum([]byte(s)))
	case "sha256":
		return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
	}
	return s
}

func (d *ExprEngine) Close() {}

// NewExprEngine returns the script engine struct
func NewExprEngine() (*ExprEngine, error) {
	logp.Debug("script", "register Expr engine")

	var err error
	d := &ExprEngine{}

	d.env = map[string]interface{}{
		"GetHEPStruct":       d.GetHEPStruct,
		"GetSIPStruct":       d.GetSIPStruct,
		"GetHEPProtoType":    d.GetHEPProtoType,
		"GetHEPSrcIP":        d.GetHEPSrcIP,
		"GetHEPSrcPort":      d.GetHEPSrcPort,
		"GetHEPDstIP":        d.GetHEPDstIP,
		"GetHEPDstPort":      d.GetHEPDstPort,
		"GetHEPTimeSeconds":  d.GetHEPTimeSeconds,
		"GetHEPTimeUseconds": d.GetHEPTimeUseconds,
		"GetHEPNodeID":       d.GetHEPNodeID,
		"GetRawMessage":      d.GetRawMessage,
		"SetRawMessage":      d.SetRawMessage,
		"SetCustomSIPHeader": d.SetCustomSIPHeader,
		"SetHEPField":        d.SetHEPField,
		"SetSIPHeader":       d.SetSIPHeader,
		"Hash":               d.Hash,
	}

	code, err := scanCode()
	if err != nil {
		return nil, err
	}

	d.prog, err = expr.Compile(code.String(), expr.Env(d.env))
	if err != nil {
		return nil, err
	}

	d.v = vm.VM{}

	return d, nil

}

// Run will execute the script
func (d *ExprEngine) Run(hep *HEP) error {
	/* preload */
	d.hepPkt = &hep

	_, err := d.v.Run(d.prog, d.env)
	if err != nil {
		return err
	}

	return nil
}
