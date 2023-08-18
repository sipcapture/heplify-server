package remotelog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/negbie/logp"
	"github.com/sipcapture/heplify-server/config"
	"github.com/sipcapture/heplify-server/decoder"
)

type Kafka struct {
	producer *kafka.Producer
	topic    string
	ctx      context.Context
}

func (k *Kafka) setup() error {
	k.ctx = context.Background()
	k.topic = config.Setting.KafkaTopic
	return k.createProducer()
}

func (k *Kafka) start(hCh chan *decoder.HEP) {
	logp.Info("start kafka...")
	go func() {
		for e := range k.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					logp.Err("Delivery failed: %v, payload: %s", ev.TopicPartition, string(ev.Value))
				}
			}
		}
	}()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case pkt, ok := <-hCh:
			if !ok {
				return
			}

			v, err := json.Marshal(pkt)
			if err != nil {
				logp.Err("%v", err)
			}
			k.producer.ProduceChannel() <- &kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &k.topic, Partition: kafka.PartitionAny},
				Value:          v,
			}
		case <-ticker.C:
			if k.producer == nil {
				k.createProducer()
			}
		}
	}
}

func (k *Kafka) createProducer() error {
	var err error
	if len(config.Setting.KafkaBroker) > 0 && len(config.Setting.KafkaTopic) > 0 {
		k.producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":   config.Setting.KafkaBroker,
			"delivery.timeout.ms": config.Setting.KafkaDeliveryTimeout,
			"linger.ms":           config.Setting.KafkaLinger,
			"batch.num.messages":  config.Setting.KafkaBatchNum,
		})
	}
	if err != nil {
		return err
	}
	return nil
}
