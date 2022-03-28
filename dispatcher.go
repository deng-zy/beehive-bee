package bee

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
)

type dispatcher struct {
	Handler
	engine    *Engine
	producer  *nsq.Producer
	snowflake *snowflake.Node
	cache     *taskCache
	db        *gorm.DB
}

var dipatcherNewHandle sync.Once
var dispatchInstance *dispatcher

func newDispatcher(e *Engine) *dispatcher {
	dipatcherNewHandle.Do(func() {
		config := nsq.NewConfig()
		deliver, err := nsq.NewProducer(e.conf.nsqd, config)
		if err != nil {
			panic(fmt.Errorf("connection to nsq fail.%w", err))
		}

		var node *snowflake.Node
		snowflake.Epoch = e.conf.snowflakeEpoch
		node, err = snowflake.NewNode(e.conf.snowflakeNode)
		if err != nil {
			e.conf.logger.Errorf("snowflake.NewNode error(%v)", err)
			node, _ = snowflake.NewNode(0)
		}

		dispatchInstance = &dispatcher{
			engine:    e,
			producer:  deliver,
			snowflake: node,
			cache:     newTaskCache(e.redis),
			db:        e.db,
		}
	})
	return dispatchInstance
}

func (d *dispatcher) Topic() string {
	return "NEW_EVENT"
}

// Concurrency 并发数
func (d *dispatcher) Concurrency() int {
	return 30
}

func (d *dispatcher) Handle(payload string) error {
	event := &Event{}
	err := json.Unmarshal([]byte(payload), event)

	if err != nil {
		return err
	}

	db := d.engine.conf.db.Begin()
	err = d.createEvent(event, db)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("create event fail(%w)", err)
	}

	var task *task
	task, err = d.createTask(event, db)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("create event fail(%w)", err)
	}

	var body []byte
	body, err = json.Marshal(task)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("marshal task fail(%w)", err)
	}

	err = d.producer.Publish(task.Topic, body)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("producer publish fail(%w)", err)
	}

	d.cache.init(task.ID)
	db.Commit()
	return nil
}

func (d *dispatcher) createEvent(m *Event, db *gorm.DB) error {
	return db.Create(m).Error
}

func (d *dispatcher) createTask(m *Event, db *gorm.DB) (*task, error) {
	task := &task{
		ID:      uint64(d.snowflake.Generate()),
		EventID: m.ID,
		Topic:   m.Topic,
		Payload: m.Payload,
		Status:  1,
	}
	err := db.Create(task).Error
	if err != nil {
		return nil, err
	}
	return task, nil
}
