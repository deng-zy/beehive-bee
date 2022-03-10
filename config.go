package bee

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Config struct {
	nsqd       string
	nsqlookupd string
	db         *gorm.DB
	logger     *logrus.Logger
}

// Option create engine option
type Option func(*Config)

// WithLogger with logger
func WithLogger(logger *logrus.Logger) Option {
	return func(c *Config) {
		c.logger = logger
	}
}

// WithDB with db
func WithDB(db *gorm.DB) Option {
	return func(c *Config) {
		c.db = db
	}
}

// WithNSQLookupd with nsqdlookup address
func WithNSQLookupd(address string) Option {
	return func(c *Config) {
		c.nsqlookupd = address
	}
}

// WithNSQd 根据nsqd地址
func WithNSQd(address string) Option {
	return func(e *Engine) {
		e.nsqd = address
	}
}

func NewConfig()
