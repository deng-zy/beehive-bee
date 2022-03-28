package bee

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewDispatch(t *testing.T) {
	eng, err := testEngine()
	if err != nil {
		t.Fatal(err)
	}

	d := newDispatcher(eng)
	if d == nil {
		t.Error("newDispather reutrn nil")
	}
}

func TestDispathHandle(t *testing.T) {
	eng, err := testEngine()
	if err != nil {
		t.Fatal(err)
	}

	d := newDispatcher(eng)
	if d == nil {
		t.Fatal("newDispather reutrn nil")
	}

	e := &Event{
		ID:          uint64(d.snowflake.Generate()),
		Topic:       "CREATE_ORDER",
		Payload:     "",
		Publisher:   "go_unit_test",
		PublishedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	payload, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("json.Marshal error. err:%v", err)
	}

	err = d.Handle(string(payload))
	if err != nil {
		t.Errorf("call displatcher.Handle return error. err:%v, payload:%s", err, payload)
	}
}
