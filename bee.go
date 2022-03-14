package bee

import (
	"encoding/json"
	"fmt"

	"github.com/nsqio/go-nsq"
)

type bee struct {
	conf     *Config
	engine   *Engine
	topic    string
	handler  IHandler
	consumer *nsq.Consumer
}

// newBee create a bee
func newBee(conf *Config, engine *Engine) *bee {
	return &bee{
		conf:   conf,
		engine: engine,
	}
}

func (b *bee) stop() {
	b.consumer.Stop()
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
		b.consumer = consumer

		consumer.AddConcurrentHandlers(b, handler.Concurrency())
		if err := consumer.ConnectToNSQLookupds(b.conf.nsqlookupd); err != nil {
			b.conf.logger.Fatal(err)
		}

		<-consumer.StopChan
	}()
}

func (b *bee) HandleMessage(m *nsq.Message) error {
	m.DisableAutoResponse()
	defer func() {
		err := recover()
		if err != nil {
			m.Requeue(-1)
		}
	}()

	var task *task
	var err error

	err = json.Unmarshal(m.Body, task)
	if err == nil {
		b.conf.logger.Errorf("unmarshal task error(%v)", err)
		m.Finish()
		return nil
	}

	err = b.handler.Handle(m.Body)
	if err == nil {
		m.Finish()
		return nil
	}

	b.conf.logger.WithField("topic", task.Topic).WithField("payload", task.Payload).Errorf("handle task fail(%v)", err)
	if !b.handler.CanRetry() {
		m.Finish()
		return nil
	}

	retires := 0
	if retires < b.handler.MaxRetries() {
		m.Requeue(-1)
		// increment retires

		// update task status
		return fmt.Errorf("task handle fail(%w)", err)
	}

	return nil
}
