// Copyright 2011, Shelby Ramsey. All rights reserved.
// Copyright 2018, Eugen Biegler. All rights reserved.
// Use of this code is governed by a BSD license that can be
// found in the LICENSE.txt file.

package sipparser

// Imports from the go standard library
import (
	"fmt"
	"strconv"
)

type RTPStatField string

const (
	RTPStatPS = "PS"
	RTPStatOS = "OS"
	RTPStatPR = "PR"
	RTPStatOR = "OR"
	RTPStatPL = "PL"
	RTPStatJI = "JI"
	RTPStatLA = "LA"
	RTPStatDU = "DU"
)

// RTPStat is a struct that holds a parsed [PX]-RTP-Stat hdr
// Fields are as follows:
// -- Val is the raw value
// -- Params has individual fields
// -- Has is a map of stat types to bool, Has[s] is true if header includes stat s.
// -- PS is the number of packets sent
// -- OS is the number of octets sent
// -- PR is the number of packets received
// -- OR is the number of octets received
// -- PL is the number of packets lost
// -- JI is the jitter
// -- LA is the round-trip delay
// -- DU is call duration
// -- Errors is a list of errors encountered when parsing individual fields
type RTPStat struct {
	Val                        string
	Params                     []*Param
	has                        map[RTPStatField]bool
	PS, OS, PR, OR, PL, JI, DU uint32
	LA                         int32
	Errors                     []error
}

// addField is a method for the Reason type that looks at the
// parameter passed into it
func (r *RTPStat) addParam(s string) {
	p := getParam(s)
	r.Params = append(r.Params, p)
	f := RTPStatField(p.Param)
	if f == RTPStatLA {
		val64, err := strconv.ParseInt(p.Val, 10, 32)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Errorf("parse err: failed to parse LA field %q as int32: %v", p.Val, err))
			return
		}
		r.LA = int32(val64)
	} else {
		val64, err := strconv.ParseUint(p.Val, 10, 32)
		if err != nil {
			r.Errors = append(r.Errors, fmt.Errorf("parse err: failed to parse %s field %q as uint32: %v", p.Param, p.Val, err))
			return
		}
		val := uint32(val64)
		switch f {
		case RTPStatPR:
			r.PR = val
		case RTPStatPS:
			r.PS = val
		case RTPStatOR:
			r.OR = val
		case RTPStatOS:
			r.OS = val
		case RTPStatPL:
			r.PL = val
		case RTPStatJI:
			r.JI = val
		case RTPStatDU:
			r.DU = val
		default:
			r.Errors = append(r.Errors, fmt.Errorf("parse err: RTP stat field %q is not supported", p.Param))
			return
		}
	}
	if r.has == nil {
		r.has = make(map[RTPStatField]bool)
	}
	r.has[f] = true
}

// parse is the method that actual parses the .Val of the RTPStat type.
func (r *RTPStat) parse() {
	for _, s := range getCommaSeperated(r.Val) {
		r.addParam(s)
	}
}

// Has returns true if the RTPStat includes stat f.
func (r *RTPStat) Has(f RTPStatField) bool {
	return r.has[f]
}
