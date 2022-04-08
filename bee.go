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
func newBee(engine *Engine, handler IHandler) (*bee, error) {
	consumer, err := newConsumer(handler.Topic())
	if err != nil {
		return nil, err
	}

	b := &bee{
		conf:     engine.conf,
		engine:   engine,
		topic:    handler.Topic(),
		handler:  handler,
		consumer: consumer,
	}

	return b, nil
}

// newConsumer create a nsq consumer
func newConsumer(topic string) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	return nsq.NewConsumer(topic, "default", config)
}

// stop stop consume nsq message
func (b *bee) stop() {
	b.consumer.Stop()
}

func (b *bee) pick() {
	defer func() {
		if err := recover(); err != nil {
			b.conf.logger.Errorf("bee.pick panic: %v", err)
			b.pick()
		}
	}()

	go func() {
		b.consumer.AddConcurrentHandlers(b, b.handler.Concurrency())
		if err := b.consumer.ConnectToNSQLookupds(b.conf.nsqlookupd); err != nil {
			b.conf.logger.Fatal(err)
		}

		<-b.consumer.StopChan
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

	if task.StartTime == nil {
		t := time.Now()
		task.StartTime = &t
	}

	if task.Status == StatusReady {
		task.Status = StatusProcessing
	}

	err = b.handler.Handle(task.Payload)
	if err == nil {
		m.Finish()
		t := time.Now()
		task.FinishTime = &t
		task.Status = StatusFinished
		task.Result = "success"
		return nil
	}

	b.conf.logger.WithField("topic", task.Topic).WithField("payload", task.Payload).Errorf("handle task fail(%v)", err)
	if !b.handler.CanRetry() {
		m.Finish()
		t := time.Now()
		task.FinishTime = &t
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

	t := time.Now()
	task.FinishTime = &t
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
