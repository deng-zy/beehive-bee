package bee

import (
	"testing"
	"time"
)

func TestNewDispatch(t *testing.T) {
	eng, err := testEngine()
	if err != nil {
		t.Fatal(err)
	}

	d, err := newDispatcher(eng)
	if d == nil {
		t.Errorf("newDispather error:%v", err)
	}
}

func TestDispathHandle(t *testing.T) {
	eng, err := testEngine()
	if err != nil {
		t.Fatalf("NewEngine error:%v", err)
	}

	d, err := newDispatcher(eng)
	if err == nil {
		t.Fatalf("newDispather error:%v", err)
	}

	e := &message{
		Topic:       "UNIT_TEST",
		Payload:     "",
		Publisher:   "go_unit_test",
		PublishedAt: time.Now(),
	}

	err = d.deliver(e)
	if err != nil {
		t.Errorf("dispather deliver task err:%v, payload:%+v", err, e)
	}
}
