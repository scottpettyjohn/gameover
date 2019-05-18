package gameover

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type (
	GameRequest struct {
		Session GameSession
	}

	GameResponse struct {
		Ok      bool
		Message string
	}

	GameSession struct {
		ExtendSession     bool
		SessionDuration   time.Duration
		remainingDuration time.Duration
	}

	GameMaster struct {
		shutdownChan    chan bool
		powerPlug       Plug
		gameOnMux       sync.Mutex
		gameOn          bool
		observers       map[Observer]struct{}
		remainingSecMux sync.Mutex
		remainingTime   time.Duration
	}
)

var gm *GameMaster

//
// Init must be called before the Game
//
func Init(requestChan <-chan GameRequest, responseChan chan<- GameResponse) *GameMaster {
	if gm != nil {
		return gm
	}
	powerPlug := PowerFactory()
	err := powerPlug.Open()
	if err != nil {
		log.Fatal(err)
	}
	gm = &GameMaster{
		shutdownChan: make(chan bool),
		powerPlug:    powerPlug,
		observers:    make(map[Observer]struct{}),
	}
	// listen for game requests and shutdown
	go func(reqChan <-chan GameRequest, respChan chan<- GameResponse) {
		for {
			request := <-reqChan
			respChan <- gm.processGameRequest(request)
		}
	}(requestChan, responseChan)
	return gm
}

//
// processGameRequest
//
func (g *GameMaster) processGameRequest(request GameRequest) GameResponse {
	g.gameOnMux.Lock()
	defer g.gameOnMux.Unlock()

	var response GameResponse
	if g.gameOn && !request.Session.ExtendSession {
		response = GameResponse{Ok: false, Message: "A game is already in progress."}
	} else if g.gameOn && request.Session.ExtendSession {
		g.remainingSecMux.Lock()
		extDuration := request.Session.SessionDuration
		g.remainingTime = g.remainingTime + extDuration
		g.remainingSecMux.Unlock()
		g.Notify(Event{Type: GameExtended, Data: int64(extDuration * time.Second)})
		response = GameResponse{Ok: true, Message: fmt.Sprintf("The current game has been extended by %s.", extDuration)}

	} else {
		success, err := g.startGame(request)
		g.gameOn = success
		if err != nil {
			response = GameResponse{Ok: false, Message: fmt.Sprintf("error when trying to start game: %s", err)}
		} else {
			response = GameResponse{Ok: true, Message: "Game on!"}
		}
	}
	return response
}

//
//
//
func (g *GameMaster) startGame(request GameRequest) (bool, error) {
	g.remainingTime = request.Session.SessionDuration
	g.powerPlug.On()
	g.Notify(Event{Type: GameStarted})
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			g.remainingSecMux.Lock()
			g.remainingTime = g.remainingTime - time.Second
			g.remainingSecMux.Unlock()
			if g.remainingTime > 0 {
				g.Notify(Event{Type: TimeRemaining, Data: int64(g.remainingTime / time.Second)})
			} else {
				g.Notify(Event{Type: TimeRemaining, Data: 0})
				ticker.Stop()
				g.powerPlug.Off()
				g.gameOnMux.Lock()
				g.gameOn = false
				g.gameOnMux.Unlock()
				g.Notify(Event{Type: GameEnded})
				return
			}
		}
	}()

	return true, nil
}

//
// Shutdown should be called before the program exits so that cleanup can occur
//
func (g *GameMaster) Shutdown() {
	g.shutdownChan <- true
	_ = g.powerPlug.Close()
}

//
// Publisher interface implementations
//
func (g *GameMaster) Register(o Observer) {
	g.observers[o] = struct{}{}
}

func (g *GameMaster) Deregister(o Observer) {
	delete(g.observers, o)
}

func (g *GameMaster) Notify(e Event) {
	for o := range g.observers {
		o.OnNotify(e)
	}
}
