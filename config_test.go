package bee

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNewConfig(t *testing.T) {
	config, err := testConfig()
	if err != nil {
		t.Error(err)
		return
	}

	if config.snowflakeNode != 2 {
		t.Errorf("snowflake nod error. except:%d", 2)
		return
	}
}

func testConnectDB() (*gorm.DB, error) {
	dsn := os.Getenv("TEST_DATABASE")
	if dsn == "" {
		return nil, errors.Errorf("require environment variable TEST_DATABASE")
	}

	return gorm.Open(mysql.Open(dsn))
}

func testConnectRedis() *redis.Client {
	host := os.Getenv("TEST_REDIS")
	if host == "" {
		host = "127.0.0.1:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr: host,
		DB:   0,
	})
	return client
}

func testNSQLOOKUP() []string {
	host := os.Getenv("TEST_NSQLOOKUP")
	if host == "" {
		return []string{"127.0.0.1:4161"}
	}
	return strings.Split(host, ",")
}

func testNSQ() string {
	host := os.Getenv("TEST_NSQ")
	if host == "" {
		return "127.0.0.1:4150"
	}
	return host
}

func testSnowflakeNode() int64 {
	node, _ := strconv.ParseInt(os.Getenv("TEST_SNOWFLAKE_NODE"), 10, 64)
	return node
}

func testSnowflakeEpoch() time.Time {
	t, _ := time.Parse("20060102150405", "20220222222222")
	return t
}

func testConfig() (*Config, error) {
	db, err := testConnectDB()
	if err != nil {
		return nil, err
	}

	redis := testConnectRedis()
	config, err := NewConfig(
		WithRedis(redis),
		WithNSQLookupd(testNSQLOOKUP()),
		WithNSQd(testNSQ()),
		WithDB(db),
		WithSnowflakeNode(testSnowflakeNode()),
		WithSnowflakeEpoch(testSnowflakeEpoch()),
	)

	if err != nil {
		return nil, err
	}

	return config, nil
}
