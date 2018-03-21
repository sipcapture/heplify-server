package queue

import (
	"strings"
	"time"

	"github.com/Shopify/sarama"
	"github.com/negbie/heplify-server/config"
	"github.com/negbie/heplify-server/logp"
)

type Kafka struct {
	producer sarama.AsyncProducer
	cfg      KafkaConfig
}

type KafkaConfig struct {
	Brokers        []string
	Compression    string
	RequestSizeMax int32
	RetryBackoff   int
	RetryMax       int
}

func (k *Kafka) setup() error {
	var (
		sCfg = sarama.NewConfig()
		err  error
	)

	k.cfg = KafkaConfig{
		Brokers:        strings.Split(config.Setting.MQAddr, ","),
		RetryMax:       3,
		RequestSizeMax: 104857600,
		RetryBackoff:   10,
	}

	sCfg.ClientID = "heplify-server-kafka"
	sCfg.Producer.Retry.Backoff = time.Duration(k.cfg.RetryBackoff) * time.Millisecond
	sCfg.Producer.Retry.Max = k.cfg.RetryMax
	sarama.MaxRequestSize = k.cfg.RequestSizeMax

	switch k.cfg.Compression {
	case "gzip":
		sCfg.Producer.Compression = sarama.CompressionGZIP
	case "lz4":
		sCfg.Producer.Compression = sarama.CompressionLZ4
	case "snappy":
		sCfg.Producer.Compression = sarama.CompressionSnappy
	default:
		sCfg.Producer.Compression = sarama.CompressionNone
	}

	k.producer, err = sarama.NewAsyncProducer(k.cfg.Brokers, sCfg)
	if err != nil {
		return err
	}

	logp.Info("kafka output address: %s, brokers: %+v, topic: %s\n", config.Setting.MQAddr, k.cfg.Brokers, config.Setting.MQTopic)

	return nil
}

func (k *Kafka) add(topic string, qCh chan []byte) {
	var (
		msg []byte
		ok  bool
	)

	for {
		msg, ok = <-qCh
		if !ok {
			break
		}

		select {
		case k.producer.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(msg),
		}:
		case err := <-k.producer.Errors():
			logp.Err("%v", err)
		}
	}

	k.producer.Close()
}
