package bee

import "github.com/nsqio/go-nsq"

type handler struct {
	topic   string
	channel chan *Event
}

func newHandler(topic string, channel ->chan *evnet) *handler {
	return &handler{
		topic: topic,
		channel, channel,
	}
}

func (h *handler) HandleMessage(message *nsq.Message) error {
	
}

func fly(e *Engine) {
	for topic, _ := range e.topics {

	}
}
