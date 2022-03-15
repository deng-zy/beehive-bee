package bee

import (
	"encoding/json"
	"fmt"
	"time"

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
		consumer.Stop()
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

	if task.StartTime.IsZero() {
		task.StartTime = time.Now()
	}

	if task.Status == StatusReady {
		task.Status = StatusProcessing
	}

	err = b.handler.Handle(task.Payload)
	defer func() {

	}()

	if err == nil {
		m.Finish()
		task.FinishTime = time.Now()
		task.Status = StatusFinished
		task.Result = "success"
		return nil
	}

	b.conf.logger.WithField("topic", task.Topic).WithField("payload", task.Payload).Errorf("handle task fail(%v)", err)
	if !b.handler.CanRetry() {
		m.Finish()
		task.FinishTime = time.Now()
		task.Status = StatusAbort
		task.Result = err.Error()
		return nil
	}

	retires := 0
	if retires < b.handler.MaxRetries() {
		m.Requeue(-1)
		task.Result = err.Error()
		return fmt.Errorf("task handle fail(%w)", err)
	}

	task.FinishTime = time.Now()
	task.Status = StatusAbort
	task.Result = err.Error()

	return nil
}

func (b *bee) retires() int {
	return 0
}

func (b *bee) markSuccess(ID uint64) {
	fields := map[string]interface{}{
		"status":      3,
		"finish_time": time.Now(),
		"result":      "success",
	}

	b.conf.db.Where("id=?", ID).Updates(fields)
}

func (b *bee) markAbort(ID uint64, err error) {
	fields := map[string]interface{}{
		"status": 4,
		"result": err.Error(),
	}

	b.conf.db.Where("id=?", ID).Updates(fields)
}
