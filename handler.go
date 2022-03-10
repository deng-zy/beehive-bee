package bee

import "encoding/json"

// IHandler event process
type IHandler interface {
	Handle([]byte) error
	CanRetry() bool
	Concurrency() int
	Topic() string
	MaxRetries() int
}

// Handler 新事件处理
type Handler struct{}

// CanRetry 是否支持重试
func (h Handler) CanRetry() bool {
	return true
}

// MaxRetries 最大重试次数
func (h Handler) MaxRetries() int {
	return 10
}

// Concurrency 并发数
func (h Handler) Concurrency() int {
	return 10
}

type eventHandler struct {
	Handler
}

func (n eventHandler) Topic() string {
	return "NEW_EVENT"
}

// Concurrency 并发数
func (n eventHandler) Concurrency() int {
	return 30
}

func (n eventHandler) Handle(payload []byte) error {
	event := &Event{}
	err := json.Unmarshal(payload, event)

	if err != nil {
		return err
	}

	return nil
}
