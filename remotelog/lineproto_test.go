package remotelog

import (
	"strings"
	"testing"
	"time"

	"github.com/sipcapture/heplify-server/decoder"
	"github.com/sipcapture/heplify-server/sipparser"
)

func TestLineprotoCreateEntry(t *testing.T) {
	// Create a test HEP packet
	hep := &decoder.HEP{
		Version:    3,
		Protocol:   17, // UDP
		SrcIP:      "192.168.1.1",
		DstIP:      "192.168.1.2",
		SrcPort:    5060,
		DstPort:    5060,
		ProtoType:  1, // SIP
		NodeID:     1,
		CID:        "a84b4c76e66710@example.com",
		Timestamp:  time.Unix(1618426800, 0),
		Payload:    "INVITE sip:user@example.com SIP/2.0\r\nVia: SIP/2.0/UDP 192.168.1.1:5060\r\n\r\n",
		SIP: &sipparser.SipMsg{
			CseqMethod: "INVITE",
		},
	}

	lp := &Lineproto{}
	entry := lp.createEntry(hep, hep.Timestamp, hep.Payload, "test-host")

	// Test measurement name
	expectedMeasurement := "hep_1"
	if entry.measurement != expectedMeasurement {
		t.Errorf("Expected measurement %s, got %s", expectedMeasurement, entry.measurement)
	}

	// Test tags
	expectedTags := map[string]string{
		"src_ip":   "192.168.1.1",
		"dst_ip":   "192.168.1.2",
		"src_port": "5060",
		"dst_port": "5060",
	}

	for k, v := range expectedTags {
		if entry.tags[k] != v {
			t.Errorf("Expected tag %s=%s, got %s", k, v, entry.tags[k])
		}
	}

	// Test fields
	if entry.fields["create_date"] != int64(1618426800000) {
		t.Errorf("Expected create_date %d, got %v", int64(1618426800000), entry.fields["create_date"])
	}

	if entry.fields["payload"] != hep.Payload {
		t.Errorf("Expected payload %s, got %s", hep.Payload, entry.fields["payload"])
	}

	if entry.fields["payload_size"] != len(hep.Payload) {
		t.Errorf("Expected payload_size %d, got %v", len(hep.Payload), entry.fields["payload_size"])
	}

	if entry.fields["sip_method"] != "INVITE" {
		t.Errorf("Expected sip_method INVITE, got %s", entry.fields["sip_method"])
	}

	if entry.fields["call_id"] != "a84b4c76e66710@example.com" {
		t.Errorf("Expected call_id %s, got %s", "a84b4c76e66710@example.com", entry.fields["call_id"])
	}
}

func TestLineprotoEncodeBatch(t *testing.T) {
	// Create test entries
	entries := []LineprotoEntry{
		{
			measurement: "hep_1",
			tags: map[string]string{
				"src_ip":   "192.168.1.1",
				"dst_ip":   "192.168.1.2",
				"src_port": "5060",
				"dst_port": "5060",
			},
			fields: map[string]interface{}{
				"create_date":   int64(1618426800000),
				"sip_method":    "INVITE",
				"call_id":       "a84b4c76e66710@example.com",
				"payload_size":  int64(245),
				"payload":       "INVITE sip:user@example.com SIP/2.0...",
			},
			timestamp: time.Unix(1618426800, 0),
		},
	}

	lp := &Lineproto{}
	buf, err := lp.encodeBatch(entries)
	if err != nil {
		t.Fatalf("Failed to encode batch: %v", err)
	}

	// Convert to string for easier testing
	output := string(buf)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(lines))
	}

	line := lines[0]
	
	// Check that the line contains expected components
	if !strings.Contains(line, "hep_1,dst_ip=192.168.1.2,dst_port=5060,src_ip=192.168.1.1,src_port=5060") {
		t.Errorf("Line missing expected tags: %s", line)
	}

	if !strings.Contains(line, "create_date=1618426800000i") {
		t.Errorf("Line missing expected create_date field: %s", line)
	}

	if !strings.Contains(line, "sip_method=\"INVITE\"") {
		t.Errorf("Line missing expected sip_method field: %s", line)
	}

	if !strings.Contains(line, "call_id=\"a84b4c76e66710@example.com\"") {
		t.Errorf("Line missing expected call_id field: %s", line)
	}

	if !strings.Contains(line, "payload_size=245i") {
		t.Errorf("Line missing expected payload_size field: %s", line)
	}

	// Check timestamp format (nanoseconds)
	if !strings.HasSuffix(line, " 1618426800000000000") {
		t.Errorf("Line missing expected timestamp: %s", line)
	}
}

func TestLineprotoEscapeFunctions(t *testing.T) {
	lp := &Lineproto{}

	// Test tag escaping
	testCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with,comma", "with\\,comma"},
		{"with space", "with\\ space"},
		{"with=equals", "with\\=equals"},
		{"with\\backslash", "with\\\\backslash"},
	}

	for _, tc := range testCases {
		result := lp.escapeTag(tc.input)
		if result != tc.expected {
			t.Errorf("escapeTag(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}

	// Test string escaping
	stringTestCases := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with\"quote", "with\\\"quote"},
		{"with\\backslash", "with\\\\backslash"},
		{"with\"quote\\backslash", "with\\\"quote\\\\backslash"},
	}

	for _, tc := range stringTestCases {
		result := lp.escapeString(tc.input)
		if result != tc.expected {
			t.Errorf("escapeString(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
} 