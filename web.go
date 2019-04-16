package gameover

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type WebSocketClient struct {
	connection *websocket.Conn
	gameMaster *GameMaster
	sync.Mutex
}

func NewWebSocketClient(conn *websocket.Conn, gm *GameMaster) *WebSocketClient {
	return &WebSocketClient{
		connection: conn,
		gameMaster: gm,
	}
}

func (w *WebSocketClient) OnNotify(e Event) {
	w.Lock()
	defer w.Unlock()
	err := w.connection.WriteJSON(e)
	if err != nil {
		log.Printf("websocket error: %s", err)
		w.connection.Close()
		w.gameMaster.Deregister(w)
	}
}
