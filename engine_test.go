package bee

import "testing"

func TestNewEngine(t *testing.T) {
	_, err := testEngine()
	if err != nil {
		t.Error(err)
	}

}

func testEngine() (*Engine, error) {
	config, err := testConfig()
	if err != nil {
		return nil, err
	}

	engine, err := NewEngine(config)
	if err != nil {
		return nil, err
	}

	return engine, nil
}
