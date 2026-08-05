package remotelog

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/negbie/logp"
	"github.com/olivere/elastic"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

type Elasticsearch struct {
	client     *elastic.Client
	bulkClient *elastic.BulkProcessor
	ctx        context.Context
}

func (e *Elasticsearch) setup() error {
	var err error
	e.ctx = context.Background()
	if len(config.Setting.ESUser) > 0 {
		e.client, err = elastic.NewClient(
			elastic.SetURL(config.Setting.ESAddr),
			elastic.SetSniff(config.Setting.ESDiscovery),
			elastic.SetBasicAuth(config.Setting.ESUser, config.Setting.ESPass),
		)
	} else {
		e.client, err = elastic.NewClient(
			elastic.SetURL(config.Setting.ESAddr),
			elastic.SetSniff(config.Setting.ESDiscovery),
		)
	}
	if err != nil {
		return err
	}

	bulkActions := config.Setting.ESBulk
	if bulkActions <= 0 {
		bulkActions = 1000
	}
	bulkSizeKB := config.Setting.ESBulkSize
	if bulkSizeKB <= 0 {
		bulkSizeKB = 2048
	}
	flushSec := config.Setting.ESTimer
	if flushSec <= 0 {
		flushSec = 10
	}
	workers := config.Setting.ESWorker
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	logp.Info("elasticsearch bulk: actions=%d size=%dKB timer=%ds workers=%d",
		bulkActions, bulkSizeKB, flushSec, workers)

	e.bulkClient, err = e.client.BulkProcessor().
		Name("ESBulkProcessor").
		Workers(workers).
		BulkActions(bulkActions).
		BulkSize(bulkSizeKB << 10).
		Stats(true).
		FlushInterval(time.Duration(flushSec) * time.Second).
		After(bulkAfter).
		Do(e.ctx)
	if err != nil {
		return err
	}

	err = showNodes(e.client)
	if err != nil {
		logp.Err("nodes info failed: %v", err)
	}

	err = e.createIndex(e.ctx, e.client)
	if err != nil {
		return err
	}

	return nil
}

func (e *Elasticsearch) start(hCh chan *decoder.HEP) {

	defer func() {
		logp.Info("heplify-server wants to stop flush remaining es bulk index requests")
		err := e.bulkClient.Flush()
		if err != nil {
			logp.Err("%v", err)
		}
	}()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				return
			}
			// Do not set .Type(): ES 8.x rejects bulk metadata with _type
			// (illegal_argument_exception: unknown parameter [_type]).
			r := elastic.NewBulkIndexRequest().Index("heplify-server-" + time.Now().Format("2006-01-02")).Doc(pkt)
			e.bulkClient.Add(r)
		case <-ticker.C:
			err := e.createIndex(e.ctx, e.client)
			if err != nil {
				logp.Warn("%v", err)
			}
		}
	}
}

func (e *Elasticsearch) createIndex(ctx context.Context, client *elastic.Client) error {
	var idx string
	// Use the IndexExists service to check if a specified index exists.
	for i := range 3 {
		t := time.Now().Add(time.Hour * time.Duration(24*i)).Format("2006-01-02")
		idx = "heplify-server-" + t
		exists, err := client.IndexExists(idx).Do(ctx)
		if err != nil {
			return err
		}
		if !exists {
			// Create a new index.
			createIndex, err := client.CreateIndex(idx).Do(ctx)
			if err != nil {
				return err
			}
			if !createIndex.Acknowledged {
				logp.Warn("creation of index %s not acknowledged", idx)
			}
			logp.Info("successfully created index %s", idx)
		} else {
			logp.Info("index %s already created", idx)
		}
	}
	return nil
}

func showNodes(client *elastic.Client) error {
	ctx := context.Background()
	info, err := client.NodesInfo().Do(ctx)
	if err != nil {
		return err
	}
	logp.Info("found cluster %q with following %d node(s)", info.ClusterName, len(info.Nodes))
	for id, node := range info.Nodes {
		logp.Info("%s, %s, %s", node.Name, id, node.TransportAddress)
	}
	return nil
}

// bulkAfter logs Elasticsearch bulk flush failures that would otherwise be silent.
func bulkAfter(_ int64, _ []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if err != nil {
		logp.Err("elasticsearch bulk flush failed: %v", err)
		return
	}
	if response == nil || !response.Errors {
		return
	}
	for _, item := range response.Failed() {
		if item != nil && item.Error != nil {
			logp.Err("elasticsearch bulk item failed: type=%s reason=%s", item.Error.Type, item.Error.Reason)
		}
	}
}

// printStats retrieves statistics from the BulkProcessor and logs them.
func printStats(e *elastic.BulkProcessor) {
	var buf bytes.Buffer
	for i, w := range e.Stats().Workers {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(fmt.Sprintf("%d=[%04d]", i, w.Queued))
	}

	logp.Info("Indexed: %05d, Succeeded: %05d, Failed: %05d, Worker Queues: %v",
		e.Stats().Indexed, e.Stats().Succeeded, e.Stats().Failed, buf.String())
}
