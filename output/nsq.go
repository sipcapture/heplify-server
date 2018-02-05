package output

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/negbie/heplify-server/logp"
	nsq "github.com/nsqio/go-nsq"
	"gopkg.in/yaml.v2"
)

type NSQ struct {
	producer *nsq.Producer
	config   NSQConfig
}

type NSQConfig struct {
	Server string `yaml:"server"`
}

func (n *NSQ) set(configFile string) error {
	var (
		err error
		cfg = nsq.NewConfig()
	)

	n.config = NSQConfig{
		Server: "localhost:4150",
	}

	cfg.UserAgent = fmt.Sprintf("heplify-server-nsq-%s", nsq.VERSION)
	cfg.DialTimeout = time.Millisecond * time.Duration(2000)

	if err = n.load(configFile); err != nil {
		logp.Err("%v", err)
	}

	n.producer, err = nsq.NewProducer(n.config.Server, cfg)
	if err != nil {
		logp.Err("%v", err)
		return err
	}
	return nil
}

func (n *NSQ) add(topic string, mCh chan []byte, ec *uint64) {
	var (
		msg []byte
		err error
		ok  bool
	)

	logp.Info("Run NSQ Output, server: %+v, topic: %s\n",
		n.config.Server, topic)

	for {
		msg, ok = <-mCh
		if !ok {
			break
		}

		err = n.producer.Publish(topic, msg)
		if err != nil {
			logp.Err("%v", err)
			*ec++
		}
	}
}

func (n *NSQ) load(f string) error {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, &n.config)
	if err != nil {
		return err
	}

	return nil
}
