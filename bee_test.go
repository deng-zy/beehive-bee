package bee

import "testing"

type testHandler struct {
	Handler
}

func (t testHandler) Concurrency() int {
	return 1
}

func (t testHandler) Handle(payload string) error {
	return nil
}

func (t testHandler) Topic() string {
	return "UNIT_TEST"
}

func TestNewBee(t *testing.T) {
	engine, err := testEngine()
	if err != nil {
		t.Errorf("NewEngine fail. err:%v", err)
		return
	}

	h := new(testHandler)
	b, err := newBee(engine, h)
	if err != nil {
		t.Errorf("newBee fail. err:%v", err)
		return
	}

	b.pick()
}
