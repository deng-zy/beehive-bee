package bee

import (
	"errors"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Config struct {
	nsqd       string
	nsqlookupd []string
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
func WithNSQLookupd(address []string) Option {
	return func(c *Config) {
		c.nsqlookupd = address
	}
}

// WithNSQd 根据nsqd地址
func WithNSQd(address string) Option {
	return func(c *Config) {
		c.nsqd = address
	}
}

// NewConfig create a config instance
func NewConfig(options ...Option) (*Config, error) {
	config := &Config{}
	for _, option := range options {
		option(config)
	}

	if config.db == nil {
		return nil, errors.New("required db")
	}

	if config.nsqd == nil || len(config.nsqd) < 1 {
		return nil, errors.New("required nsqd address")
	}

	if config.nsqlookupd == nil || len(config.nsqlookupd) < 1 {
		return nil, errors.New("required nsqlookupd address")
	}

	if config.logger == nil {
		config.logger = logrus.New()
	}

	return config, nil
}
