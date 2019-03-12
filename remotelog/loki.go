package remotelog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/decoder"
	"github.com/negbie/logp"
	"github.com/prometheus/common/model"
)

const (
	contentType  = "application/x-protobuf"
	postPath     = "/api/prom/push"
	getPath      = "/api/prom/label"
	jobName      = model.LabelValue("heplify-server")
	maxErrMsgLen = 1024
)

type entry struct {
	labels model.LabelSet
	logproto.Entry
}

type Loki struct {
	URL           string
	BatchWait     time.Duration
	BatchSize     int
	HEPTypeFilter []int
	entry
}

func (l *Loki) setup() error {
	l.BatchSize = config.Setting.LokiBulk * 1024
	l.BatchWait = time.Duration(config.Setting.LokiTimer) * time.Second
	l.HEPTypeFilter = config.Setting.LokiHEPFilter
	l.URL = config.Setting.LokiURL

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
		keep        bool
		hepType     string
		curPktTime  time.Time
		lastPktTime time.Time
		batch       = map[model.Fingerprint]*logproto.Stream{}
		batchSize   = 0
		maxWait     = time.NewTimer(l.BatchWait)
	)

	defer func() {
		if err := l.sendBatch(batch); err != nil {
			logp.Err("loki flush: %v", err)
		}
	}()

	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				return
			}
			curPktTime = pkt.Timestamp
			// guard against entry out of order errors
			if lastPktTime.After(curPktTime) {
				curPktTime = time.Now()
			}
			lastPktTime = curPktTime
			hepType = decoder.HEPTypeString(pkt.ProtoType)
			maxWait.Reset(l.BatchWait)

			pktMeta.Reset()
			pktMeta.WriteString(pkt.Payload)
			pktMeta.WriteString(" SrcIP=")
			pktMeta.WriteString(pkt.SrcIP)
			pktMeta.WriteString(" SrcPort=")
			pktMeta.WriteString(strconv.Itoa(int(pkt.SrcPort)))
			pktMeta.WriteString(" DstIP=")
			pktMeta.WriteString(pkt.DstIP)
			pktMeta.WriteString(" DstPort=")
			pktMeta.WriteString(strconv.Itoa(int(pkt.DstPort)))
			pktMeta.WriteString(" CID=")
			pktMeta.WriteString(pkt.CID)

			for _, v := range l.HEPTypeFilter {
				if pkt.ProtoType == uint32(v) {
					keep = true
					break
				}
				keep = false
			}

			switch {
			case keep && pkt.SIP != nil && pkt.ProtoType == 1:
				l.entry = entry{
					model.LabelSet{
						"job":      jobName,
						"type":     model.LabelValue(hepType),
						"node":     model.LabelValue(pkt.Node),
						"response": model.LabelValue(pkt.SIP.FirstMethod),
						"method":   model.LabelValue(pkt.SIP.CseqMethod)},
					logproto.Entry{
						Timestamp: curPktTime,
						Line:      pktMeta.String(),
					}}

			case keep && pkt.ProtoType > 1 && pkt.ProtoType <= 100:
				l.entry = entry{
					model.LabelSet{
						"job":  jobName,
						"type": model.LabelValue(hepType),
						"node": model.LabelValue(pkt.Node)},
					logproto.Entry{
						Timestamp: curPktTime,
						Line:      pktMeta.String(),
					}}
			case keep && pkt.ProtoType == 112:
				l.entry = entry{
					model.LabelSet{
						"job":  jobName,
						"type": model.LabelValue(pkt.CID),
						"node": model.LabelValue(pkt.Node),
						"host": model.LabelValue(pkt.Host)},
					logproto.Entry{
						Timestamp: curPktTime,
						Line:      pktMeta.String(),
					}}
			case pkt.ProtoType >= 1000:
				l.entry = entry{
					model.LabelSet{
						"job":  jobName,
						"type": model.LabelValue(hepType)},
					logproto.Entry{
						Timestamp: curPktTime,
						Line:      pktMeta.String(),
					}}
			default:
				continue
			}

			if batchSize+len(l.entry.Line) > l.BatchSize {
				if err := l.sendBatch(batch); err != nil {
					logp.Err("send size batch: %v", err)
				}
				batchSize = 0
				batch = map[model.Fingerprint]*logproto.Stream{}
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
