package bee

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/internal/atomic"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Engine bee engine
type Engine struct {
	handlers   []IHandler
	topics     map[string]bool
	ctx        context.Context
	channels   map[string]chan *Event
	mutex      sync.RWMutex
	logger     *logrus.Logger
	db         *gorm.DB
	running    int32
	nsqlookupd string
	nsqd       string
}

// Option create engine option
type Option func(*Engine)

// WithLogger with logger
func WithLogger(logger *logrus.Logger) Option {
	return func(e *Engine) {
		e.logger = logger
	}
}

// WithDB with db
func WithDB(db *gorm.DB) Option {
	return func(e *Engine) {
		e.db = db
	}
}

// WithNSQLookupd with nsqdlookup address
func WithNSQLookupd(address string) Option {
	return func(e *Engine) {
		e.nsqlookupd = address
	}
}

// WithNSQd 根据nsqd地址
func WithNSQd(address string) Option {
	return func(e *Engine) {
		e.nsqd = address
	}
}

// NewEngine create a new engine
func NewEngine(ctx context.Context, options ...Option) (*Engine, error) {
	e := &Engine{
		handlers:   make([]IHandler, 0, 1024),
		topics:     map[string]bool{},
		ctx:        ctx,
		channels:   map[string]chan *Event{},
		logger:     logrus.New(),
		running:    0,
		nsqlookupd: "",
		nsqd:       "",
	}

	for _, option := range options {
		option(e)
	}

	if e.nsqlookupd == "" && e.nsqd == "" {
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
	fly(e)
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
