package bee

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	OPENED = iota
	CLOSED
)

// Engine bee engine
type Engine struct {
	handlers   []IHandler
	conf       *Config
	topics     map[string]bool
	ctx        context.Context
	channels   map[string]chan *Event
	mutex      sync.RWMutex
	logger     *logrus.Logger
	db         *gorm.DB
	state      int32
	nsqlookupd string
	nsqd       string
}

// NewEngine create a new engine
func NewEngine(ctx context.Context, conf *Config) (*Engine, error) {
	e := &Engine{
		handlers: make([]IHandler, 0, 1024),
		topics:   map[string]bool{},
		ctx:      ctx,
		channels: map[string]chan *Event{},
	}

	if e.nsqlookupd == "" || e.nsqd == "" {
		return nil, errors.New("you have to setup a nsqlookupd or nsqd")
	}

	if e.db == nil {
		return nil, errors.New("required db connection")
	}

	return e, nil
}

// AddHandler 添加
func (e *Engine) AddHandler(h IHandler) {
	e.mutex.RLock()
	topic := h.Topic()
	_, exists := e.topics[topic]
	e.mutex.RUnlock()

	if exists {
		return
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.topics[topic] = true
	e.handlers = append(e.handlers, h)
	e.channels[topic] = make(chan *Event, h.Concurrency())
}

// Run start event handle
func (e *Engine) Run() {
	if !atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		return
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-c:
		break
	case <-e.ctx.Done():
		break
	}

	atomic.CompareAndSwapInt32(&e.running, 1, 0)
}
