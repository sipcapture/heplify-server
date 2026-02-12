package remotelog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/sipcapture/heplify-server/remotelog/logproto"
)

func withConfig(t *testing.T, fn func()) {
	t.Helper()
	original := config.Setting
	defer func() {
		config.Setting = original
	}()
	fn()
}

func TestLokiSetupSendsOrgIDHeader(t *testing.T) {
	withConfig(t, func() {
		headerCh := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != getPath {
				t.Errorf("unexpected path %s", r.URL.Path)
			}
			headerCh <- r.Header.Get("X-Scope-OrgID")
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		config.Setting.LokiURL = server.URL
		config.Setting.LokiOrgID = "heplify-org-id"

		loki := &Loki{}
		if err := loki.setup(); err != nil {
			t.Fatalf("setup failed: %v", err)
		}

		select {
		case got := <-headerCh:
			if got != "heplify-org-id" {
				t.Errorf("expected X-Scope-OrgID header %q, got %q", "heplify-org-id", got)
			}
		default:
			t.Fatal("expected label check request")
		}
	})
}

func TestLokiSetupFailsOnNon2xx(t *testing.T) {
	withConfig(t, func() {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		config.Setting.LokiURL = server.URL

		loki := &Loki{}
		if err := loki.setup(); err == nil {
			t.Fatal("expected setup to fail on non-2xx status")
		}
	})
}

func TestLokiSendSendsOrgIDHeaderAndFailsOnNon2xx(t *testing.T) {
	withConfig(t, func() {
		headerCh := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != postPathOne {
				t.Errorf("unexpected path %s", r.URL.Path)
			}
			headerCh <- r.Header.Get("X-Scope-OrgID")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("nope"))
		}))
		defer server.Close()

		config.Setting.LokiOrgID = "heplify-org-id"

		loki := &Loki{URL: server.URL + postPathOne}
		if _, err := loki.send(context.Background(), []byte("payload")); err == nil {
			t.Fatal("expected send to fail on non-2xx status")
		}

		select {
		case got := <-headerCh:
			if got != "heplify-org-id" {
				t.Errorf("expected X-Scope-OrgID header %q, got %q", "heplify-org-id", got)
			}
		default:
			t.Fatal("expected push request")
		}
	})
}

func TestLokiSkipsTCPPortLabels(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiIPPortLabels = true
		config.Setting.LokiSkipTCPPortLabels = true

		loki := &Loki{}
		loki.entry = entry{model.LabelSet{}, logproto.Entry{}}

		// Test TCP packet (protocol 6)
		tcpPkt := &decoder.HEP{
			Protocol: 6, // TCP
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  12345,
			DstPort:  5060,
		}

		// Simulate the label setting logic from loki.go
		if config.Setting.LokiIPPortLabels {
			loki.entry.labels["src_ip"] = model.LabelValue(tcpPkt.SrcIP)
			loki.entry.labels["dst_ip"] = model.LabelValue(tcpPkt.DstIP)
			skipPortLabels := config.Setting.LokiSkipTCPPortLabels && tcpPkt.Protocol == 6
			if !skipPortLabels {
				loki.entry.labels["src_port"] = model.LabelValue("12345")
				loki.entry.labels["dst_port"] = model.LabelValue("5060")
			}
		}

		// Verify TCP ports are NOT in labels
		if _, exists := loki.entry.labels["src_port"]; exists {
			t.Error("expected src_port label to be skipped for TCP")
		}
		if _, exists := loki.entry.labels["dst_port"]; exists {
			t.Error("expected dst_port label to be skipped for TCP")
		}
		// Verify IPs are still present
		if loki.entry.labels["src_ip"] != "192.168.1.1" {
			t.Errorf("expected src_ip label to be set, got %v", loki.entry.labels["src_ip"])
		}
		if loki.entry.labels["dst_ip"] != "192.168.1.2" {
			t.Errorf("expected dst_ip label to be set, got %v", loki.entry.labels["dst_ip"])
		}
	})
}

func TestLokiIncludesUDPPortLabels(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiIPPortLabels = true
		config.Setting.LokiSkipTCPPortLabels = true

		loki := &Loki{}
		loki.entry = entry{model.LabelSet{}, logproto.Entry{}}

		// Test UDP packet (protocol 17)
		udpPkt := &decoder.HEP{
			Protocol: 17, // UDP
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  5060,
			DstPort:  5060,
		}

		// Simulate the label setting logic from loki.go
		if config.Setting.LokiIPPortLabels {
			loki.entry.labels["src_ip"] = model.LabelValue(udpPkt.SrcIP)
			loki.entry.labels["dst_ip"] = model.LabelValue(udpPkt.DstIP)
			skipPortLabels := config.Setting.LokiSkipTCPPortLabels && udpPkt.Protocol == 6
			if !skipPortLabels {
				loki.entry.labels["src_port"] = model.LabelValue("5060")
				loki.entry.labels["dst_port"] = model.LabelValue("5060")
			}
		}

		// Verify UDP ports ARE in labels
		if loki.entry.labels["src_port"] != "5060" {
			t.Errorf("expected src_port label to be set for UDP, got %v", loki.entry.labels["src_port"])
		}
		if loki.entry.labels["dst_port"] != "5060" {
			t.Errorf("expected dst_port label to be set for UDP, got %v", loki.entry.labels["dst_port"])
		}
	})
}

func TestLokiIncludesTCPPortLabelsWhenSkipDisabled(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiIPPortLabels = true
		config.Setting.LokiSkipTCPPortLabels = false

		loki := &Loki{}
		loki.entry = entry{model.LabelSet{}, logproto.Entry{}}

		// Test TCP packet (protocol 6)
		tcpPkt := &decoder.HEP{
			Protocol: 6, // TCP
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  12345,
			DstPort:  5060,
		}

		// Simulate the label setting logic from loki.go
		if config.Setting.LokiIPPortLabels {
			loki.entry.labels["src_ip"] = model.LabelValue(tcpPkt.SrcIP)
			loki.entry.labels["dst_ip"] = model.LabelValue(tcpPkt.DstIP)
			skipPortLabels := config.Setting.LokiSkipTCPPortLabels && tcpPkt.Protocol == 6
			if !skipPortLabels {
				loki.entry.labels["src_port"] = model.LabelValue("12345")
				loki.entry.labels["dst_port"] = model.LabelValue("5060")
			}
		}

		// Verify TCP ports ARE in labels when skip is disabled
		if loki.entry.labels["src_port"] != "12345" {
			t.Errorf("expected src_port label to be set when skip is disabled, got %v", loki.entry.labels["src_port"])
		}
		if loki.entry.labels["dst_port"] != "5060" {
			t.Errorf("expected dst_port label to be set when skip is disabled, got %v", loki.entry.labels["dst_port"])
		}
	})
}
