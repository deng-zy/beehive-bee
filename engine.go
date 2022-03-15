package bee

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	// StatusReady task status ready
	StatusReady = 1
	// StatusProcessing task status processing
	StatusProcessing = 2
	// StatusFinished task status Finished
	StatusFinished = 3
	// StatusAbort task status abort
	StatusAbort = 4
	// StatusCancel task status cancel
	StatusCancel = 5
)

// Engine bee engine
type Engine struct {
	handlers map[string]IHandler
	conf     *Config
	mutex    sync.RWMutex
	bees     []*bee
	db       *gorm.DB
	redis    *redis.Client
}

// NewEngine create a new engine
func NewEngine(conf *Config) (*Engine, error) {
	e := &Engine{
		handlers: map[string]IHandler{},
		conf:     conf,
		db:       conf.db,
		redis:    conf.redis,
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
