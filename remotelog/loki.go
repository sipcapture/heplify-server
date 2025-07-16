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
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/negbie/logp"
	"github.com/prometheus/common/model"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
	"github.com/sipcapture/heplify-server/remotelog/logproto"
)

const (
	contentType  = "application/x-protobuf"
	postPath     = "/api/prom/push"
	postPathOne  = "/loki/api/v1/push"
	getPath      = "/loki/api/v1/label"
	jobName      = model.LabelValue("heplify-server")
	maxErrMsgLen = 1024
)

type entry struct {
	labels model.LabelSet
	logproto.Entry
}

type Loki struct {
	URL             string
	BatchWait       time.Duration
	BatchSize       int
	AllowOutOfOrder bool
	entry
}

func (l *Loki) setup() error {
	l.BatchSize = config.Setting.LokiBulk * 1024
	l.BatchWait = time.Duration(config.Setting.LokiTimer) * time.Second
	l.URL = config.Setting.LokiURL
	l.AllowOutOfOrder = config.Setting.LokiAllowOutOfOrder

	u, err := url.Parse(l.URL)
	if err != nil {
		return err
	}
	if !strings.Contains(u.Path, postPath) {
		u.Path = postPath
		q := u.Query()
		u.RawQuery = q.Encode()
		l.URL = u.String()
	}
	l.URL = strings.Replace(l.URL, postPath, postPathOne, -1)
	u.Path = getPath
	q := u.Query()
	u.RawQuery = q.Encode()

	_, err = http.Get(u.String())
	if err != nil {
		return err
	}
	return nil
}

func (l *Loki) start(hCh chan *decoder.HEP) {
	var (
		pktMeta     strings.Builder
		curPktTime  time.Time
		lastPktTime time.Time
		batch       = map[model.Fingerprint]*logproto.Stream{}
		batchSize   = 0
		maxWait     = time.NewTimer(l.BatchWait)
		hostname    string
	)

	defer func() {
		if err := l.sendBatch(batch); err != nil {
			logp.Err("loki flush: %v", err)
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
			if !l.AllowOutOfOrder {
				// guard against entry out of order errors
				if lastPktTime.After(curPktTime) {
					curPktTime = time.Now()
				}
				lastPktTime = curPktTime
			}

			pktMeta.Reset()

			l.entry = entry{model.LabelSet{}, logproto.Entry{Timestamp: curPktTime}}

			if pkt.ProtoString == "rtcp" {
				var document map[string]interface{}
				err := json.Unmarshal([]byte(pkt.Payload), &document)
				if err != nil {
					logp.Err("Unable to decode rtcp json: %v", err)
					pktMeta.WriteString(pkt.Payload)
				} else {
					document["cid"] = pkt.CID
					documentJson, err := json.Marshal(document)
					if config.Setting.LokiCallIDLabels {
						l.entry.labels["call_id"] = model.LabelValue(pkt.CID)
					}
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

			switch {
			case pkt.SIP != nil && pkt.ProtoType == 1:
				l.entry.labels["method"] = model.LabelValue(pkt.SIP.CseqMethod)
				l.entry.labels["response"] = model.LabelValue(pkt.SIP.FirstMethod)
				protocol := ""
				if pkt.Protocol == 6 {
					protocol = "tcp"
				} else if pkt.Protocol == 17 {
					protocol = "udp"
				}
				l.entry.labels["protocol"] = model.LabelValue(protocol)
				if config.Setting.LokiCallIDLabels {
					l.entry.labels["call_id"] = model.LabelValue(pkt.SIP.CallID)
				}
				if config.Setting.LokiFromToLabels {
					l.entry.labels["from"] = model.LabelValue(pkt.SIP.From.Val)
					l.entry.labels["to"] = model.LabelValue(pkt.SIP.To.Val)
				}
			case pkt.ProtoType == 100:
				protocol := "udp"
				if strings.Contains(pkt.Payload, "Fax") || strings.Contains(pkt.Payload, "T38") {
					protocol = "fax"
				} else if strings.Contains(pkt.Payload, "sip") {
					protocol = "sip"
				}
				l.entry.labels["protocol"] = model.LabelValue(protocol)

			}

			l.entry.labels["job"] = jobName
			l.entry.labels["hostname"] = model.LabelValue(hostname)
			l.entry.labels["node"] = model.LabelValue(pkt.NodeName)
			l.entry.labels["type"] = model.LabelValue(pkt.ProtoString)
			l.entry.Entry.Line = pktMeta.String()

			if config.Setting.LokiIPPortLabels {
				l.entry.labels["src_ip"] = model.LabelValue(pkt.SrcIP)
				l.entry.labels["src_port"] = model.LabelValue(strconv.FormatUint(uint64(pkt.SrcPort), 10))
				l.entry.labels["dst_ip"] = model.LabelValue(pkt.DstIP)
				l.entry.labels["dst_port"] = model.LabelValue(strconv.FormatUint(uint64(pkt.DstPort), 10))
			}

			if batchSize+len(l.entry.Line) > l.BatchSize {
				if err := l.sendBatch(batch); err != nil {
					logp.Err("send size batch: %v", err)
				}
				batchSize = 0
				batch = map[model.Fingerprint]*logproto.Stream{}
				maxWait.Reset(l.BatchWait)
			}

			batchSize += len(l.entry.Line)
			fp := l.entry.labels.FastFingerprint()
			stream, ok := batch[fp]
			if !ok {
				stream = &logproto.Stream{
					Labels: l.entry.labels.String(),
				}
				batch[fp] = stream
			}
			stream.Entries = append(stream.Entries, l.Entry)

		case <-maxWait.C:
			if len(batch) > 0 {
				if err := l.sendBatch(batch); err != nil {
					logp.Err("send time batch: %v", err)
				}
				batchSize = 0
				batch = map[model.Fingerprint]*logproto.Stream{}
			}
			maxWait.Reset(l.BatchWait)
		}
	}
}

func (l *Loki) sendBatch(batch map[model.Fingerprint]*logproto.Stream) error {
	buf, err := encodeBatch(batch)
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

func encodeBatch(batch map[model.Fingerprint]*logproto.Stream) ([]byte, error) {
	req := logproto.PushRequest{
		Streams: make([]*logproto.Stream, 0, len(batch)),
	}
	for _, stream := range batch {
		req.Streams = append(req.Streams, stream)
	}
	buf, err := proto.Marshal(&req)
	if err != nil {
		return nil, err
	}
	buf = snappy.Encode(nil, buf)
	return buf, nil
}

func (l *Loki) send(ctx context.Context, buf []byte) (int, error) {
	req, err := http.NewRequest("POST", l.URL, bytes.NewReader(buf))
	if err != nil {
		return -1, err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", contentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	logp.Debug("loki", "%s request with %d bytes to %s - %v response", req.Method, len(buf), l.URL, resp.StatusCode)

	if resp.StatusCode/100 != 2 {
		scanner := bufio.NewScanner(io.LimitReader(resp.Body, maxErrMsgLen))
		line := ""
		if scanner.Scan() {
			line = scanner.Text()
		}
		err = fmt.Errorf("server returned HTTP status %s (%d): %s", resp.Status, resp.StatusCode, line)
	}
	return resp.StatusCode, err
}
