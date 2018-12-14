package remotelog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/olivere/elastic"
)

type Elasticsearch struct {
	client     *elastic.Client
	bulkClient *elastic.BulkProcessor
	ctx        context.Context
}

func (e *Elasticsearch) setup() error {
	var err error
	e.ctx = context.Background()
	for {
		e.client, err = elastic.NewClient(
			elastic.SetURL(config.Setting.ESAddr),
			elastic.SetSniff(config.Setting.ESDiscovery),
		)
		if err != nil {
			logp.Err("%v", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	e.bulkClient, err = e.client.BulkProcessor().
		Name("ESBulkProcessor").
		Workers(runtime.NumCPU()).
		BulkActions(1000).
		BulkSize(2 << 20).
		Stats(true).
		FlushInterval(10 * time.Second).
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

	go func(e *elastic.BulkProcessor) {
		for range time.Tick(5 * time.Minute) {
			printStats(e)
		}
	}(e.bulkClient)

	return nil
}

func (e *Elasticsearch) send(hCh chan *decoder.HEP) {
	var (
		pkt *decoder.HEP
		ok  bool
	)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(12 * time.Hour)

	for {
		select {
		case pkt, ok = <-hCh:
			if !ok {
				break
			}
			r := elastic.NewBulkIndexRequest().Index("heplify-server-" + time.Now().Format("2006-01-02")).Type("hep").Doc(pkt)
			e.bulkClient.Add(r)
		case <-ticker.C:
			err := e.createIndex(e.ctx, e.client)
			if err != nil {
				logp.Warn("%v", err)
			}
		case <-c:
			logp.Info("heplify-server wants to stop flush remaining es bulk index requests")
			e.bulkClient.Flush()
		}
	}
}

func (e *Elasticsearch) createIndex(ctx context.Context, client *elastic.Client) error {
	var idx string
	// Use the IndexExists service to check if a specified index exists.
	for i := 0; i < 3; i++ {
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
