package remotelog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes/timestamp"
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
	getPath      = "/api/prom/label"
	jobName      = model.LabelValue("heplify-server")
	maxErrMsgLen = 1024
)

type entry struct {
	labels model.LabelSet
	*logproto.Entry
}

type Loki struct {
	URL       string
	BatchWait time.Duration
	BatchSize int
	entry
}

func (l *Loki) setup() error {
	l.BatchSize = config.Setting.LokiBulk * 1024
	l.BatchWait = time.Duration(config.Setting.LokiTimer) * time.Second
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

			pktMeta.Reset()
			pktMeta.WriteString(pkt.Payload)
			pktMeta.WriteString(" src_ip=")
			pktMeta.WriteString(pkt.SrcIP)
			pktMeta.WriteString(" dst_ip=")
			pktMeta.WriteString(pkt.DstIP)
			if pkt.ProtoType < 110 {
				pktMeta.WriteString(" id=")
				pktMeta.WriteString(pkt.CID)
			}

			tsNano := curPktTime.UnixNano()
			ts := &timestamp.Timestamp{
				Seconds: tsNano / int64(time.Second),
				Nanos:   int32(tsNano % int64(time.Second)),
			}

			switch {
			case pkt.SIP != nil && pkt.ProtoType == 1:
				l.entry = entry{
					model.LabelSet{
						"job":      jobName,
						"type":     model.LabelValue(pkt.ProtoString),
						"node":     model.LabelValue(pkt.NodeName),
						"response": model.LabelValue(pkt.SIP.FirstMethod),
						"method":   model.LabelValue(pkt.SIP.CseqMethod)},
					&logproto.Entry{
						Timestamp: ts,
						Line:      pktMeta.String(),
					}}

			case pkt.ProtoType > 1 && pkt.ProtoType <= 100:
				l.entry = entry{
					model.LabelSet{
						"job":  jobName,
						"type": model.LabelValue(pkt.ProtoString),
						"node": model.LabelValue(pkt.NodeName)},
					&logproto.Entry{
						Timestamp: ts,
						Line:      pktMeta.String(),
					}}
			case pkt.ProtoType == 112:
				l.entry = entry{
					model.LabelSet{
						"job":   jobName,
						"level": model.LabelValue(pkt.CID),
						"node":  model.LabelValue(pkt.NodeName),
						"host":  model.LabelValue(pkt.HostTag)},
					&logproto.Entry{
						Timestamp: ts,
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
