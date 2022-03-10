package bee

import (
	"log"

	"github.com/nsqio/go-nsq"
)

type bee struct {
	conf    *Config
	engine  *Engine
	topic   string
	handler IHandler
}

// newBee create a bee
func newBee(conf *Config, engine *Engine) *bee {
	return &bee{
		conf:   conf,
		engine: engine,
	}
}

func (b *bee) pick(topic string, handler IHandler) {
	b.topic = topic
	b.handler = handler

	go func() {
		config := nsq.NewConfig()
		consumer, err := nsq.NewConsumer(topic, "", config)
		if err != nil {
			b.conf.logger.Fatalf("new %s consumer fail. error:%v", topic, err)
		}
		consumer.AddConcurrentHandlers(b, handler.Concurrency())
		if err := consumer.ConnectToNSQLookupds(b.conf.nsqlookupd); err != nil {
			log.Fatal(err)
		}
		<-consumer.StopChan
	}()
}

func (b *bee) HandleMessage(m *nsq.Message) error {
	err := b.handler.Handle(m.Body)
	if err == nil {
		return nil
	}

	return nil
}
