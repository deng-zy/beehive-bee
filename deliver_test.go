package bee

import "testing"

func TestDeliverPublish(t *testing.T) {
	nsq := testNSQ()
	d, err := NewDeliver(nsq)
	if err != nil {
		t.Fatalf("NewDeliver error:%v", err)
	}

	err = d.Publish("UNIT_TEST", "UNIT_TEST", "goUnitTest")
	if err != nil {
		t.Errorf("Deliver publish error:%v", err)
	}
}
