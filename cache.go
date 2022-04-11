package bee

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	statusField     = "status"
	retryField      = "retires"
	startTimeField  = "start_time"
	finishTimeField = "finish_time"
	resultField     = "result"
)

// taskCache 事件处理任务缓存
type taskCache struct {
	redis     *redis.Client
	ctx       context.Context
	keyPrefix string
	expire    time.Duration
	fields    []string
}

// cacheInstance 事件处理环缓存实例
var cacheInstance *taskCache

// cacheHandle 事件处理缓存执行器
var cacheHandle sync.Once

// newTaskCache 创建新的事件处理缓存
func newTaskCache(client *redis.Client) *taskCache {
	cacheHandle.Do(func() {
		cacheInstance = &taskCache{
			keyPrefix: "t:",
			expire:    168 * time.Hour,
			redis:     client,
			ctx:       context.Background(),
			fields:    []string{statusField, retryField, startTimeField, finishTimeField, resultField},
		}
	})

	return cacheInstance
}

// init 初始事件处理缓存
func (t *taskCache) init(ID int64) {
	key := t.key(ID)
	t.redis.HSet(t.ctx, key, statusField, StatusProcessing, resultField, "", retryField, 0, startTimeField, 0, finishTimeField, 0)
}

// get 获取事件处理缓存数据
func (t *taskCache) get(ID int64) map[string]string {
	key := t.key(ID)
	cache, _ := t.redis.HGetAll(t.ctx, key).Result()
	return cache
}

// success 更新事件处理完成时需要更新缓存的数据
func (t *taskCache) success(ID int64, finishedAt time.Time) {
	key := t.key(ID)
	p := t.redis.Pipeline()

	p.HSet(t.ctx, key, statusField, StatusFinished, finishTimeField, finishedAt.Format(time.RFC3339Nano), resultField, "success")
	p.Expire(t.ctx, key, t.expire)
	p.Exec(t.ctx)
}

// retrying 更新事件处理失败继续重试时需要更新的数据
func (t *taskCache) retrying(ID int64, result error) {
	p := t.redis.Pipeline()
	key := t.key(ID)
	p.HSet(t.ctx, key, resultField, result.Error())
	p.HIncrBy(t.ctx, key, retryField, 1)
	p.Exec(t.ctx)
}

// abort 更新事件终止继续时需要更新的数据
func (t *taskCache) abort(ID int64, abortedAt time.Time, result error) {
	p := t.redis.Pipeline()
	key := t.key(ID)
	p.HSet(t.ctx, key, statusField, StatusAbort, finishTimeField, abortedAt.Format(time.RFC3339Nano), resultField, result.Error())
	p.Expire(t.ctx, key, t.expire)
	p.Exec(t.ctx)
}

// retires 获取重试次数
func (t *taskCache) retires(ID int64) int {
	val, _ := t.redis.HGet(t.ctx, t.key(ID), retryField).Result()
	retires, _ := strconv.ParseInt(val, 10, 32)
	return int(retires)
}

// setStartTime 设置事件处理开始时间
func (t *taskCache) setStartTime(ID int64, startTime time.Time) {
	t.redis.HSet(t.ctx, t.key(ID), startTimeField, startTime)
}

// key 获取事件缓存key
func (t *taskCache) key(ID int64) string {
	return t.keyPrefix + strconv.FormatInt(ID, 10)
}
