// Copyright 2011, Shelby Ramsey. All rights reserved.
// Copyright 2018, Eugen Biegler. All rights reserved.
// Use of this code is governed by a BSD license that can be
// found in the LICENSE.txt file.

package sipparser

// Imports from the go standard library

/*
func TestRTPStat(t *testing.T) {
	sm := &SipMsg{}
	s := "PS=295,OS=5900,PR=292,OR=5840,PL=0,JI=0,LA=0,DU=5,XX=123"
	sm.parseRTPStat(s)
	if got := len(sm.RTPStat.Errors); got != 1 {
		t.Errorf("parseRTPStat: got %d parse errors: %q, want one - unsupported XX", got, sm.RTPStat.Errors)
	}
	wantFields := []RTPStatField{"PS", "OS", "PR", "OR", "PL", "JI", "LA", "DU"}
	var gotFields []RTPStatField
	for f, _ := range sm.RTPStat.has {
		gotFields = append(gotFields, f)
	}
	if got, want := len(gotFields), len(wantFields); got != want {
		t.Errorf("RTPStat number of parsed fields: got %d, want %d", got, want)
	}
	for _, f := range wantFields {
		if !sm.RTPStat.Has(f) {
			t.Errorf("RTPStat.Has(%q): got false, want true", f)
		}
	}
	if got, want := sm.RTPStat.PS, uint32(295); got != want {
		t.Errorf("RTPStat.PS: got %d, want %d", got, want)
	}
	if got, want := sm.RTPStat.LA, int32(0); got != want {
		t.Errorf("RTPStat.LA: got %d, want %d", got, want)
	}
}
*/
