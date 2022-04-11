package bee

import (
	"encoding/json"
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/nsqio/go-nsq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// DispatherTopic dispather listen queue name
const DispatherTopic = "NEW_EVENET"

type dispatcher struct {
	Handler
	engine    *Engine
	producer  *nsq.Producer
	snowflake *snowflake.Node
	cache     *taskCache
	db        *gorm.DB
	consumer  *nsq.Consumer
}

var dispatcherHandle sync.Once
var dispatcherInst *dispatcher

func newDispatcher(e *Engine) (*dispatcher, error) {
	var err error
	dispatcherHandle.Do(func() {
		var consumer *nsq.Consumer
		consumer, err = newConsumer(DispatherTopic)
		if err != nil {
			return
		}

		var producer *nsq.Producer
		producer, err = newProducer(e.conf.nsqd)
		if err != nil {
			return
		}

		var node *snowflake.Node
		node, err = newSnowflake(e.conf.snowflakeNode, e.conf.snowflakeEpoch)

		if err != nil {
			return
		}

		dispatcherInst = &dispatcher{
			engine:    e,
			producer:  producer,
			snowflake: node,
			cache:     newTaskCache(e.redis),
			db:        e.db,
			consumer:  consumer,
		}
	})

	if err != nil {
		return nil, err
	}
	return dispatcherInst, nil
}

func (d *dispatcher) listen() {
	go func() {
		d.consumer.AddConcurrentHandlers(d, d.Concurrency())
		<-d.consumer.StopChan
	}()
}

func (d *dispatcher) stop() {
	d.consumer.Stop()
}

func (d *dispatcher) HandleMessage(m *nsq.Message) error {
	m.DisableAutoResponse()

	msg := &message{}
	if err := json.Unmarshal(m.Body, msg); err != nil {
		d.engine.conf.logger.Errorf("Unmarshal event fail. error:%v", err)
		return nil
	}

	defer func() {
		if err := recover(); err != nil {
			m.Requeue(-1)
		} else {
			m.Finish()
		}
	}()

	err := d.deliver(msg)
	if err != nil {
		return errors.Wrap(err, "create task error")
	}

	return nil
}

func (d *dispatcher) deliver(m *message) error {
	db := d.engine.conf.db.Begin()
	var err error

	defer func() {
		if p := recover(); p != nil {
			db.Rollback()
		} else if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	e := &Event{
		ID:          d.snowflake.Generate().Int64(),
		Topic:       m.Topic,
		Payload:     m.Payload,
		Publisher:   m.Publisher,
		PublishedAt: m.PublishedAt,
	}

	err = db.Create(e).Error
	if err != nil {
		return err
	}

	t := &task{
		ID:      d.snowflake.Generate().Int64(),
		EventID: e.ID,
		Topic:   e.Topic,
		Payload: e.Payload,
		Status:  StatusReady,
	}

	err = db.Create(t).Error
	if err != nil {
		return err
	}

	body, err := json.Marshal(t)
	if err != nil {
		return err
	}

	d.producer.Publish(t.Topic, body)
	d.cache.init(t.ID)
	return nil
}

// Topic return dispatcher nsq topic
func (d *dispatcher) Topic() string {
	return DispatherTopic
}

// Concurrency 并发数
func (d *dispatcher) Concurrency() int {
	return 30
}
