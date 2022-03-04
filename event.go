package bee

// Event event
type Event struct {
	ID      uint64 `json:"id"`
	UUID    string `json:"uuid"`
	Topic   string `json:"topic"`
	Message string `json:"message"`
}
