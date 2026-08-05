package remotelog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/common/model"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
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

func TestApplyLokiIPPortLabels_SkipsTCPPortsByDefault(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiSkipTCPPortLabels = true
		labels := model.LabelSet{}
		applyLokiIPPortLabels(&labels, &decoder.HEP{
			Protocol: 6,
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  12345,
			DstPort:  5060,
		})
		if _, ok := labels["src_port"]; ok {
			t.Fatal("expected src_port to be skipped for TCP")
		}
		if _, ok := labels["dst_port"]; ok {
			t.Fatal("expected dst_port to be skipped for TCP")
		}
		if labels["src_ip"] != "192.168.1.1" || labels["dst_ip"] != "192.168.1.2" {
			t.Fatalf("unexpected IP labels: %v", labels)
		}
	})
}

func TestApplyLokiIPPortLabels_IncludesUDPPorts(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiSkipTCPPortLabels = true
		labels := model.LabelSet{}
		applyLokiIPPortLabels(&labels, &decoder.HEP{
			Protocol: 17,
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  5060,
			DstPort:  5060,
		})
		if labels["src_port"] != "5060" || labels["dst_port"] != "5060" {
			t.Fatalf("expected UDP ports in labels, got %v", labels)
		}
	})
}

func TestApplyLokiIPPortLabels_IncludesTCPPortsWhenSkipDisabled(t *testing.T) {
	withConfig(t, func() {
		config.Setting.LokiSkipTCPPortLabels = false
		labels := model.LabelSet{}
		applyLokiIPPortLabels(&labels, &decoder.HEP{
			Protocol: 6,
			SrcIP:    "192.168.1.1",
			DstIP:    "192.168.1.2",
			SrcPort:  12345,
			DstPort:  5060,
		})
		if labels["src_port"] != "12345" || labels["dst_port"] != "5060" {
			t.Fatalf("expected TCP ports when skip disabled, got %v", labels)
		}
	})
}

