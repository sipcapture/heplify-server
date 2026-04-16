package decoder

import (
	"strings"
	"testing"

	"github.com/sipcapture/heplify-server/config"
)

func newTestEngine(hep *HEP) *LuaEngine {
	return &LuaEngine{hepPkt: &hep}
}

func TestSetLokiLabel_Accepted(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{"tenant_id"}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("tenant_id", "acme")

	if hep.CustomLokiLabels["tenant_id"] != "acme" {
		t.Errorf("expected label to be set, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_NotAllowlisted(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{"tenant_id"}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("other_key", "value")

	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected no labels set, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_NoAllowlistConfigured(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("tenant_id", "acme")

	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected no labels when allowlist is empty, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_InvalidKey(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{"bad-key"}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("bad-key", "value")

	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected key with hyphen to be rejected, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_ReservedKey(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{"job"}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("job", "custom")

	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected reserved key to be rejected, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_ValueTooLong(t *testing.T) {
	config.Setting.LokiCustomLabels = []string{"tenant_id"}
	hep := &HEP{}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("tenant_id", strings.Repeat("x", lokiLabelMaxValLen+1))

	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected oversized value to be rejected, got %v", hep.CustomLokiLabels)
	}
}

func TestSetLokiLabel_MaxLabelsExceeded(t *testing.T) {
	keys := []string{"k1", "k2", "k3", "k4", "k5", "k6"}
	config.Setting.LokiCustomLabels = keys
	hep := &HEP{}
	eng := newTestEngine(hep)

	for _, k := range keys {
		eng.SetLokiLabel(k, "v")
	}

	if len(hep.CustomLokiLabels) != lokiLabelMaxCount {
		t.Errorf("expected exactly %d labels, got %d", lokiLabelMaxCount, len(hep.CustomLokiLabels))
	}
}

func TestSetLokiLabel_BackwardCompatibility(t *testing.T) {
	// When LokiCustomLabels is unset, existing packets are unaffected.
	config.Setting.LokiCustomLabels = nil
	hep := &HEP{Payload: "original"}
	eng := newTestEngine(hep)

	eng.SetLokiLabel("tenant_id", "acme")

	if hep.Payload != "original" {
		t.Error("existing HEP fields must not be modified")
	}
	if len(hep.CustomLokiLabels) != 0 {
		t.Errorf("expected no custom labels, got %v", hep.CustomLokiLabels)
	}
}
