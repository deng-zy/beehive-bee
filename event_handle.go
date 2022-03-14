package bee

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nsqio/go-nsq"
	"gorm.io/gorm"
)

type eventHandler struct {
	Handler
	engine  *Engine
	deliver *nsq.Producer
}

var ehOnce sync.Once
var ehi *eventHandler //
func newEventHandler(e *Engine) *eventHandler {
	ehOnce.Do(func() {
		config := nsq.NewConfig()
		deliver, err := nsq.NewProducer(e.conf.nsqd, config)
		if err != nil {
			panic(fmt.Errorf("connection to nsq fail.%w", err))
		}

		ehi = &eventHandler{
			engine:  e,
			deliver: deliver,
		}
	})
	return ehi
}

func (n eventHandler) Topic() string {
	return "NEW_EVENT"
}

// Concurrency 并发数
func (n eventHandler) Concurrency() int {
	return 30
}

func (n eventHandler) Handle(payload []byte) error {
	event := &Event{}
	err := json.Unmarshal(payload, event)

	if err != nil {
		return err
	}

	db := n.engine.conf.db.Begin()
	err = n.createEvent(event, db)
	if err != nil {
		db.Rollback()
		return fmt.Errorf("create event fail(%w)", err)
	}

	var task *Task
	task, err = n.createTask(event, db)
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

	err = n.deliver.Publish(task.Topic, body)
	if err == nil {
		db.Commit()
		return nil
	}

	db.Rollback()
	return fmt.Errorf("producer publish fail(%w)", err)
}

func (n eventHandler) createEvent(m *Event, db *gorm.DB) error {
	return db.Create(m).Error
}

func (n eventHandler) createTask(m *Event, db *gorm.DB) (*Task, error) {
	task := &Task{
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
