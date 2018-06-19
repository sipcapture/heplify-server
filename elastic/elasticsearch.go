package elastic

import (
	"context"
	"runtime"
	"time"

	"github.com/negbie/heplify-server"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/logp"
	"github.com/olivere/elastic"
)

type Elasticsearch struct {
	bulkClient *elastic.BulkProcessor
}

func (e *Elasticsearch) setup() error {
	var err error
	var client *elastic.Client
	ctx := context.Background()
	for {
		client, err = elastic.NewClient(
			elastic.SetURL(config.Setting.ESAddr),
			elastic.SetSniff(false),
		)
		if err != nil {
			logp.Err("%v", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}
	}
	e.bulkClient, err = client.BulkProcessor().
		Name("ESBulkProcessor").
		Workers(runtime.NumCPU()).
		BulkActions(2000).
		BulkSize(2 << 20).
		FlushInterval(10 * time.Second).
		Do(ctx)
	if err != nil {
		return err
	}
	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("heplify-server").Do(ctx)
	if err != nil {
		return err
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex("heplify-server").Do(ctx)
		if err != nil {
			return err
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}
	return nil
}

func (e *Elasticsearch) send(hCh chan *decoder.HEP) {
	var (
		pkt *decoder.HEP
		ok  bool
	)

	for {
		pkt, ok = <-hCh
		if !ok {
			break
		}
		r := elastic.NewBulkIndexRequest().Index("heplify-server").Type("hep").Doc(pkt)
		e.bulkClient.Add(r)
	}
}
