package bee

import (
	"errors"
	"strconv"
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	e, err := testEngine()
	if err != nil {
		t.Fatal(err)
	}

	cache := newTaskCache(e.redis)
	if cache == nil {
		t.Fatal("newTaskCache fail. instance nil")
	}

	var task uint64 = 1
	cache.init(task)

	//test setStartTime
	now := time.Now()
	cache.setStartTime(task, now)
	i := cache.get(task)
	st, err := time.Parse(time.RFC3339, i["start_time"])
	if err != nil {
		t.Errorf("task startTime error. time:%s, err:%v.", i["start_time"], err)
		return
	}

	if st.UnixMilli() != now.UnixMilli() {
		t.Errorf("task startTime error. execept:%s, actual:%s", now.Format(time.RFC3339), st.Format(time.RFC3339))
	}

	//test retrying
	cache.retrying(task, errors.New("unit test retrying"))
	i = cache.get(task)
	if cache.retires(task) != 1 {
		t.Errorf("incrment task retires fail. cache:%v", i)
		return
	}

	//test abort
	abort := errors.New("unit test abort")
	cache.abort(task, time.Now(), abort)
	i = cache.get(task)
	if i["result"] != abort.Error() {
		t.Errorf("task cache abort reason error. cache:%v, except:%s, acutal:%s", i, abort, i["result"])
		return
	}

	if i["status"] != strconv.FormatInt(StatusAbort, 10) {
		t.Errorf("task cache abort status error. cache:%v, except:%d, acutal:%s", i, StatusAbort, i["status"])
		return
	}

	//test success
	cache.success(task, time.Now())
	i = cache.get(task)
	if i["status"] != strconv.FormatInt(StatusFinished, 10) {
		t.Errorf("task cache status error. cache:%v, except:%d, acutal:%s", i, StatusFinished, i["status"])
		return
	}

	if i["result"] != "success" {
		t.Errorf("task cache result error. cache:%v, except:%s, acutal:%s", i, "success", i["result"])
		return
	}

	t.Logf("cache info:%v", i)
}
