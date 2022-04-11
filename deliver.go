package bee

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/nsqio/go-nsq"
)

// Deliver event deliver
type Deliver struct {
	addr     string
	producer *nsq.Producer
}

// message deliver message
type message struct {
	Topic       string    `json:"topic"`
	Payload     string    `json:"payload"`
	Publisher   string    `json:"publisher"`
	PublishedAt time.Time `json:"published_at"`
}

var deliverHandle sync.Once
var deliverInstance *Deliver

// NewDeliver create a new deliver instance
func NewDeliver(addr string) (*Deliver, error) {
	var err error
	var producer *nsq.Producer
	deliverHandle.Do(func() {
		conf := nsq.NewConfig()
		producer, err = nsq.NewProducer(addr, conf)
		if err != nil {
			return
		}

		deliverInstance = &Deliver{
			addr:     addr,
			producer: producer,
		}
	})

	if err != nil {
		return nil, err
	}

	return deliverInstance, nil
}

// Publish publish message to nsq
func (d *Deliver) Publish(topic, payload, publisher string) error {
	m := &message{
		Topic:       topic,
		Payload:     payload,
		Publisher:   publisher,
		PublishedAt: time.Now(),
	}

	body, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return d.producer.Publish(DispatherTopic, body)
}
