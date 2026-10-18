package remotelog

import (
	"testing"

	"github.com/prometheus/common/model"
	"github.com/sipcapture/heplify-server/remotelog/logproto"
)

// 486 to 172.17.47.195 and 487 to 172.17.47.194 share a FastFingerprint.
// The other labels are the ones heplify puts on an INVITE response.
func collidingLabelSets(t *testing.T) (model.LabelSet, model.LabelSet) {
	t.Helper()
	common := model.LabelSet{
		"hostname": "heplify-server-2",
		"job":      "heplify-server",
		"method":   "INVITE",
		"node":     "0",
		"protocol": "tcp",
		"src_ip":   "172.17.34.196",
		"type":     "sip",
	}
	busy := model.LabelSet{
		"dst_ip":   "172.17.47.195",
		"response": "486",
	}
	terminated := model.LabelSet{
		"dst_ip":   "172.17.47.194",
		"response": "487",
	}
	for k, v := range common {
		busy[k] = v
		terminated[k] = v
	}
	if busy.FastFingerprint() != terminated.FastFingerprint() {
		t.Fatal("fixture no longer collides under FastFingerprint")
	}
	return busy, terminated
}

func TestLokiStreamSeparatesFastFingerprintCollision(t *testing.T) {
	busy, terminated := collidingLabelSets(t)
	batch := map[model.Fingerprint]*logproto.Stream{}

	busyStream := lokiStream(batch, busy)
	terminatedStream := lokiStream(batch, terminated)
	if busyStream == terminatedStream {
		t.Fatal("colliding label sets shared one stream")
	}
	if busyStream.Labels != busy.String() {
		t.Fatalf("busy stream labels = %s, want %s", busyStream.Labels, busy.String())
	}
	if terminatedStream.Labels != terminated.String() {
		t.Fatalf("terminated stream labels = %s, want %s", terminatedStream.Labels, terminated.String())
	}
	if lokiStream(batch, busy) != busyStream || lokiStream(batch, terminated) != terminatedStream {
		t.Fatal("a later packet did not rejoin its own stream")
	}
	if len(batch) != 2 {
		t.Fatalf("batch has %d streams, want 2", len(batch))
	}
}
