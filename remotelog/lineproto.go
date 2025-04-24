package remotelog

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

const (
	contentType = "text/plain"
	postPath    = "/write"
)

type LineProto struct {
	URL       string
	BatchWait time.Duration
	BatchSize int
}

func (l *LineProto) setup() error {
	l.BatchSize = config.Setting.LineProtoBulk
	l.BatchWait = time.Duration(config.Setting.LineProtoTimer) * time.Second
	l.URL = config.Setting.LineProtoURL 
	u, err := url.Parse(l.URL)
	if err != nil {
		return err
	}
	
	u.Path = postPath
	q := u.Query()
	q.Set("db", "hep")
	u.RawQuery = q.Encode()
	l.URL = u.String()

	// Test connection
	_, err = http.Get(l.URL)
	if err != nil {
		return err
	}
	return nil
}

func (l *LineProto) start(hCh chan *decoder.HEP) {
	var (
		batch     strings.Builder
		batchSize = 0
		maxWait   = time.NewTimer(l.BatchWait)
		hostname  string
	)

	defer func() {
		if err := l.sendBatch(batch.String()); err != nil {
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

			// Create measurement name based on HEP type
			measurement := fmt.Sprintf("hep_%d", pkt.ProtoType)

			// Build tags
			tags := []string{
				fmt.Sprintf("src_ip=%s", pkt.SrcIP),
				fmt.Sprintf("dst_ip=%s", pkt.DstIP),
				fmt.Sprintf("src_port=%d", pkt.SrcPort),
				fmt.Sprintf("dst_port=%d", pkt.DstPort),
				fmt.Sprintf("hostname=%s", hostname),
				fmt.Sprintf("node=%s", pkt.NodeName),
			}

			// Add protocol tag
			protocol := "unknown"
			if pkt.Protocol == 6 {
				protocol = "tcp"
			} else if pkt.Protocol == 17 {
				protocol = "udp"
			}
			tags = append(tags, fmt.Sprintf("protocol=%s", protocol))

			// Build fields
			fields := []string{
				fmt.Sprintf("create_date=%di", pkt.Timestamp.UnixNano()/1e6), // Convert to milliseconds
				fmt.Sprintf("payload_size=%di", len(pkt.Payload)),
			}

			// Add SIP specific fields if available
			if pkt.SIP != nil && pkt.ProtoType == 1 {
				fields = append(fields,
					fmt.Sprintf("sip_method=%q", pkt.SIP.CseqMethod),
					fmt.Sprintf("sip_response=%q", pkt.SIP.FirstMethod),
				)
				if pkt.SIP.CallID != "" {
					fields = append(fields, fmt.Sprintf("call_id=%q", pkt.SIP.CallID))
				}
			}

			// Add payload as a field
			fields = append(fields, fmt.Sprintf("payload=%q", pkt.Payload))

			// Construct the line protocol entry
			line := fmt.Sprintf("%s,%s %s %d\n",
				measurement,
				strings.Join(tags, ","),
				strings.Join(fields, ","),
				pkt.Timestamp.UnixNano(),
			)

			if batchSize+len(line) > l.BatchSize {
				if err := l.sendBatch(batch.String()); err != nil {
					logp.Err("send size batch: %v", err)
				}
				batch.Reset()
				batchSize = 0
				maxWait.Reset(l.BatchWait)
			}

			batch.WriteString(line)
			batchSize += len(line)

		case <-maxWait.C:
			if batchSize > 0 {
				if err := l.sendBatch(batch.String()); err != nil {
					logp.Err("send time batch: %v", err)
				}
				batch.Reset()
				batchSize = 0
			}
			maxWait.Reset(l.BatchWait)
		}
	}
}

func (l *LineProto) sendBatch(batch string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequest("POST", l.URL, strings.NewReader(batch))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	logp.Debug("lineproto", "POST request with %d bytes to %s - %v response", len(batch), l.URL, resp.StatusCode)

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("server returned HTTP status %s (%d)", resp.Status, resp.StatusCode)
	}
	return nil
} 
