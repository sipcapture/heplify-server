package remotelog

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/logproto"
	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/prometheus/common/model"
)

const contentType = "application/x-protobuf"

type entry struct {
	labels model.LabelSet
	logproto.Entry
}

type Loki struct {
	URL       string
	BatchWait time.Duration
	BatchSize int
	quit      chan struct{}
	entry
	wg sync.WaitGroup
}

func (l *Loki) setup() error {
	l.BatchSize = config.Setting.LokiBulk
	l.BatchWait = time.Duration(config.Setting.LokiTimer) * time.Second
	l.URL = config.Setting.LokiURL
	l.quit = make(chan struct{})

	return nil
}

func (l *Loki) send(hCh chan *decoder.HEP) {
	var (
		pkt *decoder.HEP
		ok  bool
	)

	batch := map[model.Fingerprint]*logproto.Stream{}
	batchSize := 0
	maxWait := time.NewTimer(l.BatchWait)

	defer func() {
		if err := l.sendBatch(batch); err != nil {
			logp.Err("send %v", err)
		}
		l.wg.Done()
	}()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(12 * time.Hour)

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}

			if pkt.SIP != nil && pkt.ProtoType == 1 {

				maxWait.Reset(l.BatchWait)
				l.entry = entry{model.LabelSet{"method": model.LabelValue(pkt.SIP.CseqVal)}, logproto.Entry{
					Timestamp: pkt.Timestamp,
					Line:      pkt.Payload,
				}}

				if batchSize+len(l.entry.Line) > l.BatchSize {
					if err := l.sendBatch(batch); err != nil {
						logp.Err("sendBatch %v", err)
					}
					batch = map[model.Fingerprint]*logproto.Stream{}
				}

				fp := l.entry.labels.FastFingerprint()
				stream, ok := batch[fp]
				if !ok {
					stream = &logproto.Stream{
						Labels: l.entry.labels.String(),
					}
					batch[fp] = stream
				}
				stream.Entries = append(stream.Entries, l.Entry)
			}
		case <-ticker.C:

		case <-c:
			logp.Info("heplify-server wants to stop flush remaining es bulk index requests")

		case <-l.quit:
			return

		}
	}
}

func (l *Loki) sendBatch(batch map[model.Fingerprint]*logproto.Stream) error {
	req := logproto.PushRequest{
		Streams: make([]*logproto.Stream, 0, len(batch)),
	}
	count := 0
	for _, stream := range batch {
		req.Streams = append(req.Streams, stream)
		count += len(stream.Entries)
	}
	buf, err := proto.Marshal(&req)
	if err != nil {
		return err
	}
	buf = snappy.Encode(nil, buf)

	resp, err := http.Post(l.URL, contentType, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Error doing write: %d - %s", resp.StatusCode, resp.Status)
	}
	return nil
}

// Stop the client.
func (l *Loki) Stop() {
	close(l.quit)
	l.wg.Wait()
}
