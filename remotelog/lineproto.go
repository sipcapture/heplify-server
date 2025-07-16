package remotelog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	lineprotoContentType = "application/octet-stream"
	lineprotoPostPath    = "/api/v2/write"
	lineprotoJobName     = "heplify-server"
	lineprotoMaxErrMsgLen = 1024
)

type LineprotoEntry struct {
	measurement string
	tags        map[string]string
	fields      map[string]interface{}
	timestamp   time.Time
}

type Lineproto struct {
	URL             string
	BatchWait       time.Duration
	BatchSize       int
	IPPortLabels    bool
	entries         []LineprotoEntry
}

func (l *Lineproto) setup() error {
	l.BatchSize = config.Setting.LineprotoBulk * 1024
	l.BatchWait = time.Duration(config.Setting.LineprotoTimer) * time.Second
	l.URL = config.Setting.LineprotoURL
	l.IPPortLabels = config.Setting.LineprotoIPPortLabels

	u, err := url.Parse(l.URL)
	if err != nil {
		return err
	}
	if !strings.Contains(u.Path, lineprotoPostPath) {
		u.Path = lineprotoPostPath
		q := u.Query()
		u.RawQuery = q.Encode()
		l.URL = u.String()
	}

	// Test connection by making a simple GET request to the base URL
	baseURL := u.String()
	baseURL = strings.Replace(baseURL, lineprotoPostPath, "/ping", -1)
	_, err = http.Get(baseURL)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lineproto) start(hCh chan *decoder.HEP) {
	var (
		pktMeta     strings.Builder
		curPktTime  time.Time
		batch       []LineprotoEntry
		batchSize   = 0
		maxWait     = time.NewTimer(l.BatchWait)
		hostname    string
	)

	defer func() {
		if err := l.sendBatch(batch); err != nil {
			logp.Err("lineproto flush: %v", err)
		}
	}()

	hostname, err := os.Hostname()
	if err != nil {
		logp.Warn("Unable to obtain hostname: %v", err)
	}

	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				return
			}
			curPktTime = pkt.Timestamp

			pktMeta.Reset()

			if pkt.ProtoString == "rtcp" {
				var document map[string]interface{}
				err := json.Unmarshal([]byte(pkt.Payload), &document)
				if err != nil {
					logp.Err("Unable to decode rtcp json: %v", err)
					pktMeta.WriteString(pkt.Payload)
				} else {
					document["cid"] = pkt.CID
					documentJson, err := json.Marshal(document)
					if err != nil {
						logp.Err("Unable to re-generate rtcp json: %v", err)
						pktMeta.WriteString(pkt.Payload)
					} else {
						pktMeta.Write(documentJson)
					}
				}
			} else {
				pktMeta.WriteString(pkt.Payload)
			}

			entry := l.createEntry(pkt, curPktTime, pktMeta.String(), hostname)

			if batchSize+len(pktMeta.String()) > l.BatchSize {
				if err := l.sendBatch(batch); err != nil {
					logp.Err("send size batch: %v", err)
				}
				batchSize = 0
				batch = []LineprotoEntry{}
				maxWait.Reset(l.BatchWait)
			}

			batchSize += len(pktMeta.String())
			batch = append(batch, entry)

		case <-maxWait.C:
			if len(batch) > 0 {
				if err := l.sendBatch(batch); err != nil {
					logp.Err("send time batch: %v", err)
				}
				batchSize = 0
				batch = []LineprotoEntry{}
			}
			maxWait.Reset(l.BatchWait)
		}
	}
}

func (l *Lineproto) createEntry(pkt *decoder.HEP, timestamp time.Time, payload string, hostname string) LineprotoEntry {
	entry := LineprotoEntry{
		measurement: fmt.Sprintf("hep_%d", pkt.ProtoType),
		tags:        make(map[string]string),
		fields:      make(map[string]interface{}),
		timestamp:   timestamp,
	}

	// Set IP/Port tags (always included as per example)
	entry.tags["src_ip"] = pkt.SrcIP
	entry.tags["dst_ip"] = pkt.DstIP
	entry.tags["src_port"] = strconv.FormatUint(uint64(pkt.SrcPort), 10)
	entry.tags["dst_port"] = strconv.FormatUint(uint64(pkt.DstPort), 10)

	// Set fields
	entry.fields["create_date"] = timestamp.UnixMilli() // milliseconds timestamp as integer
	entry.fields["payload"] = payload
	entry.fields["payload_size"] = len(payload)

	// Add SIP-specific fields when available
	if pkt.SIP != nil && pkt.ProtoType == 1 {
		if pkt.SIP.CseqMethod != "" {
			entry.fields["sip_method"] = pkt.SIP.CseqMethod
		}
		if pkt.CID != "" {
			entry.fields["call_id"] = pkt.CID
		}
	}

	return entry
}

func (l *Lineproto) sendBatch(batch []LineprotoEntry) error {
	if len(batch) == 0 {
		return nil
	}

	buf, err := l.encodeBatch(batch)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = l.send(ctx, buf)
	if err != nil {
		return err
	}
	return nil
}

func (l *Lineproto) encodeBatch(batch []LineprotoEntry) ([]byte, error) {
	var buf strings.Builder

	for _, entry := range batch {
		// Build the line protocol format: measurement,tag1=value1,tag2=value2 field1=value1,field2=value2 timestamp
		buf.WriteString(entry.measurement)

		// Add tags (sorted alphabetically)
		if len(entry.tags) > 0 {
			buf.WriteString(",")
			tagKeys := make([]string, 0, len(entry.tags))
			for k := range entry.tags {
				tagKeys = append(tagKeys, k)
			}
			sort.Strings(tagKeys)
			tagPairs := make([]string, 0, len(entry.tags))
			for _, k := range tagKeys {
				v := entry.tags[k]
				escapedKey := l.escapeTag(k)
				escapedValue := l.escapeTag(v)
				tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", escapedKey, escapedValue))
			}
			buf.WriteString(strings.Join(tagPairs, ","))
		}

		// Add fields
		buf.WriteString(" ")
		fieldPairs := make([]string, 0, len(entry.fields))
		for k, v := range entry.fields {
			escapedKey := l.escapeField(k)
			switch val := v.(type) {
			case string:
				// Use WriteJSONString for proper escaping of payload and other string fields
				var fieldBuf strings.Builder
				fieldBuf.WriteString(escapedKey)
				fieldBuf.WriteString("=\"")
				decoder.WriteJSONString(&fieldBuf, val)
				fieldBuf.WriteString("\"")
				fieldPairs = append(fieldPairs, fieldBuf.String())
			case int, int32, int64, uint, uint32, uint64:
				fieldPairs = append(fieldPairs, fmt.Sprintf("%s=%di", escapedKey, val))
			case float32, float64:
				fieldPairs = append(fieldPairs, fmt.Sprintf("%s=%f", escapedKey, val))
			case bool:
				fieldPairs = append(fieldPairs, fmt.Sprintf("%s=%t", escapedKey, val))
			default:
				// Convert to string for unknown types
				var fieldBuf strings.Builder
				fieldBuf.WriteString(escapedKey)
				fieldBuf.WriteString("=\"")
				decoder.WriteJSONString(&fieldBuf, fmt.Sprintf("%v", val))
				fieldBuf.WriteString("\"")
				fieldPairs = append(fieldPairs, fieldBuf.String())
			}
		}
		buf.WriteString(strings.Join(fieldPairs, ","))

		// Add timestamp (nanoseconds)
		buf.WriteString(fmt.Sprintf(" %d\n", entry.timestamp.UnixNano()))
	}

	return []byte(buf.String()), nil
}

func (l *Lineproto) escapeTag(s string) string {
	// Escape commas, spaces, and equals signs in tag keys and values
	s = strings.ReplaceAll(s, "\\", "\\\\") // escape backslash first
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, " ", "\\ ")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

func (l *Lineproto) escapeField(s string) string {
	// Escape spaces in field keys
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, " ", "\\ ")
	s = strings.ReplaceAll(s, "=", "\\=")
	return s
}

func (l *Lineproto) escapeString(s string) string {
	// Escape backslash first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func (l *Lineproto) send(ctx context.Context, buf []byte) (int, error) {
	req, err := http.NewRequest("POST", l.URL, bytes.NewReader(buf))
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", lineprotoContentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	logp.Debug("lineproto", "%s request with %d bytes to %s - %v response", req.Method, len(buf), l.URL, resp.StatusCode)

	if resp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, lineprotoMaxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s (%d): %s", resp.Status, resp.StatusCode, line)
	}
	return resp.StatusCode, err
} 