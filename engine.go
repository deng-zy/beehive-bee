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
}

// NewEngine create a new engine
func NewEngine(conf *Config) (*Engine, error) {
	e := &Engine{
		handlers: map[string]IHandler{},
		conf:     conf,
	}

	e.AddHandler(new(eventHandler))

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
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	<-c
}
