// Copyright 2011, Shelby Ramsey. All rights reserved.
// Copyright 2018, Eugen Biegler. All rights reserved.
// Use of this code is governed by a BSD license that can be
// found in the LICENSE.txt file.

package sipparser

// Imports from the go standard library
import (
	"testing"
)

func TestUri(t *testing.T) {
	s := "sip:15555551000;npdi=yes;rn=15555551999@0.0.0.0:5060;user=phone;key"
	u := ParseURI(s)
	if u.Scheme != SIP_SCHEME {
		t.Errorf("[TestUri] Error parsing URI \"sip:15555551000;npdi=yes;rn=15555551999@0.0.0.0:5060;user=phone;key\".  Scheme should be sip not received val: " + u.Scheme)
	}
	if u.User != "15555551000" {
		t.Errorf("[TestUri] Error parsing URI \"sip:15555551000@0.0.0.0:5060;user=phone\".  User value is not \"15555551000\".")
	}
	if u.Host != "0.0.0.0" {
		t.Errorf("[TestUri] Error parsing URI \"sip:15555551000@0.0.0.0:5060;user=phone\".  Host value should be \"0.0.0.0\" but received: " + u.Host)
	}
	if u.Port != "5060" {
		t.Errorf("[TestUri] Error parsing URI \"sip:15555551000@0.0.0.0:5060;user=phone\". Port value should be '5060' ... but it is not.")
	}
	if u.PortInt != 5060 {
		t.Errorf("[TestUri] Error parsing URI \"sip:15555551000@0.0.0.0:5060;user=phone\". Port value should be 5060 ... but it is not.")
	}
	s = "tel:5554448000@myfoo.com"
	u = ParseURI(s)
	if u.User != "5554448000" {
		t.Errorf("[TestUri] Error parsing URI \"tel:5554448000@myfoo.com\".  Should have received \"5554448000\" as the user.  Received: " + u.User)
	}
	if u.Host != "myfoo.com" {
		t.Errorf("[TestUri] Error parsing URI \"tel:5554448000@myfoo.com\".  Host should be \"myfoo.com\" but received: " + u.Host)
	}
	s = "tel:+5554448000"
	u = ParseURI(s)
	if u.User != "+5554448000" {
		t.Errorf("[TestUri] Error parsing URI \"tel:+5554448000\".  Should have received \"+5554448000\" as the user.  Received: " + u.User)
	}
	s = "sip:myfoo.com"
	u = ParseURI(s)
	if u.Raw != "myfoo.com" {
		t.Errorf("[TestUri] Error parsing URI \"sip:myfoo.com\".  Should have received \"myfoo.com\" as Raw.  Received: " + u.Raw)
	}
	if u.User != "" {
		t.Errorf("[TestUri] Error parsing URI \"sip:myfoo.com\".  User should be \"\" but received: " + u.User)
	}
	if u.Host != "myfoo.com" {
		t.Errorf("[TestUri] Error parsing URI \"sip:myfoo.com\".  Host should be \"myfoo.com\" but received: " + u.Host)
	}
}

func TestUriIPv6Bracketed(t *testing.T) {
	u := ParseURI("sip:+32460214964@[2a0c:10c0:ffff:2::1234]")
	if u.Error != nil {
		t.Fatalf("unexpected error: %v", u.Error)
	}
	if got, want := u.Host, "2a0c:10c0:ffff:2::1234"; got != want {
		t.Fatalf("Host = %q, want %q", got, want)
	}
	if u.Port != "" {
		t.Fatalf("Port = %q, want empty", u.Port)
	}
}

func TestUriIPv6BracketedWithPort(t *testing.T) {
	u := ParseURI("sip:user@[2001:db8::1]:5060")
	if u.Error != nil {
		t.Fatalf("unexpected error: %v", u.Error)
	}
	if got, want := u.Host, "2001:db8::1"; got != want {
		t.Fatalf("Host = %q, want %q", got, want)
	}
	if got, want := u.Port, "5060"; got != want {
		t.Fatalf("Port = %q, want %q", got, want)
	}
	if u.PortInt != 5060 {
		t.Fatalf("PortInt = %d, want 5060", u.PortInt)
	}
}

func TestUriIPv6Unbracketed(t *testing.T) {
	u := ParseURI("sip:9876@fe80:0:0:0:21a:a0ff:fe07:32e0")
	if u.Error != nil {
		t.Fatalf("unexpected error: %v", u.Error)
	}
	if got, want := u.Host, "fe80:0:0:0:21a:a0ff:fe07:32e0"; got != want {
		t.Fatalf("Host = %q, want %q", got, want)
	}
}

func TestFromHeaderIPv6Domain(t *testing.T) {
	s := ParseMsg("INVITE sip:a@b SIP/2.0\r\nFrom: \"+32460214964\" <sip:+32460214964@[2a0c:10c0:ffff:2::1234]>;tag=x\r\nTo: <sip:b@c>\r\nCall-ID: 1\r\nCSeq: 1 INVITE\r\n\r\n", nil, nil)
	if got, want := s.FromHost, "2a0c:10c0:ffff:2::1234"; got != want {
		t.Fatalf("FromHost = %q, want %q", got, want)
	}
}
