package bee

import (
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Config bee运行时配置信息
type Config struct {
	nsqd           string
	nsqlookupd     []string
	snowflakeEpoch int64
	snowflakeNode  int64
	db             *gorm.DB
	logger         *logrus.Logger
	redis          *redis.Client
}

// Option create a new config
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

// WithSnowflakeNode snowflake id node
func WithSnowflakeNode(node int64) Option {
	return func(c *Config) {
		c.snowflakeNode = node
	}
}

// WithSnowflakeEpoch snowflake epoch
func WithSnowflakeEpoch(t time.Time) Option {
	return func(c *Config) {
		c.snowflakeEpoch = t.UnixMilli()
	}
}

// WithRedis redis client
func WithRedis(client *redis.Client) Option {
	return func(c *Config) {
		c.redis = client
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

	if config.redis == nil {
		return nil, errors.New("required redis")
	}

	if config.nsqd == "" {
		return nil, errors.New("required nsqd address")
	}

	if config.nsqlookupd == nil || len(config.nsqlookupd) < 1 {
		return nil, errors.New("required nsqlookupd address")
	}

	if config.logger == nil {
		config.logger = logrus.New()
	}

	if config.snowflakeEpoch == 0 {
		t, _ := time.Parse("2006-01-02 15:04:05", "2022-02-22 22:22:22")
		config.snowflakeEpoch = t.UnixMilli()
	}

	return config, nil
}
