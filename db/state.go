package db

type StatusPool struct {
	messages map[int64]*StatusMessage
}

func (sp *StatusPool) SetStatus(message StatusMessage) {
	sp.messages[message.Value] = &message
}

func (sp *StatusPool) PopStatus(value int64) (*StatusMessage, bool) {
	m, ok := sp.messages[value]
	if ok {
		delete(sp.messages, value)
	}
	return m, ok
}

func NewStatusPool() *StatusPool {
	newStatus := &StatusPool{messages: make(map[int64]*StatusMessage)}
	newStatus.SetStatus(StatusMessage{Value: 0, Message: "Database initialization started"})
	newStatus.SetStatus(StatusMessage{Value: 1, Message: "Setting up database connection"})
	newStatus.SetStatus(StatusMessage{Value: 2, Message: "Initializing database"})
	newStatus.SetStatus(StatusMessage{Value: 3, Message: "Loading kanji dictionary"})
	newStatus.SetStatus(StatusMessage{Value: 4, Message: "Loading jmdict dictionary"})
	newStatus.SetStatus(StatusMessage{Value: 5, Message: "Loading tatoeba sentences"})

	return &StatusPool{}
}

type StatusMessage struct {
	Value   int64
	Message string
}
