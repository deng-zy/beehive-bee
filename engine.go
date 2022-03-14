package bee

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Engine bee engine
type Engine struct {
	handlers map[string]IHandler
	conf     *Config
	mutex    sync.RWMutex
	bees     []*bee
}

// NewEngine create a new engine
func NewEngine(conf *Config) (*Engine, error) {
	e := &Engine{
		handlers: map[string]IHandler{},
		conf:     conf,
		bees:     make([]*bee, 1024),
	}

	h := newEventHandler(e)
	e.AddHandler(h)

	return e, nil
}

// AddHandler 添加
func (e *Engine) AddHandler(h IHandler) {
	e.mutex.RLock()
	topic := h.Topic()
	_, exists := e.handlers[topic]
	e.mutex.RUnlock()

	if exists {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.handlers[topic] = h
}

// Run start event handle
func (e *Engine) Run() {
	e.mutex.RLock()
	for topic, handler := range e.handlers {
		bee := newBee(e.conf, e)
		e.bees = append(e.bees, bee)
		bee.pick(topic, handler)
	}
	e.mutex.RUnlock()

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	<-c

	for _, bee := range e.bees {
		bee.stop()
	}
}
