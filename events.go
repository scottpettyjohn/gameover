package gameover

type EventType int

const (
	GameStarted EventType = iota
	GamePaused
	GameEnded
	TimeRemaining
	GameExtended
)

type (
	Event struct {
		Type EventType
		Data int64
	}

	Observer interface {
		OnNotify(Event)
	}

	Publisher interface {
		Register(Observer)
		Deregister(Observer)
		Notify(Event)
	}
)
