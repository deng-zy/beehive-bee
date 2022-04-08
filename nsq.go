package bee

import "github.com/nsqio/go-nsq"

// newConsumer create a nsq consumer
func newConsumer(topic string) (*nsq.Consumer, error) {
	config := nsq.NewConfig()
	return nsq.NewConsumer(topic, "default", config)
}

// newProducer create a nsq producer
func newProducer(addr string) (*nsq.Producer, error) {
	config := nsq.NewConfig()
	return nsq.NewProducer(addr, config)
}
