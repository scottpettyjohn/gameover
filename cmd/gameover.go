package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/scottpettyjohn/gameover"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type PlaySession struct {
	Duration string `json:"duration"`
}

func (ps PlaySession) parseDuration() time.Duration {
	return 0
}

var (
	gameMaster     *gameover.GameMaster
	gameRequestCh  chan gameover.GameRequest  = make(chan gameover.GameRequest)
	gameResponseCh chan gameover.GameResponse = make(chan gameover.GameResponse)
	upgrader                                  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {

	gameMaster = gameover.Init(gameRequestCh, gameResponseCh)
	gameMaster.Register(&gameover.Foley{})
	log.Println("ready player 1.")
	router := mux.NewRouter()
	router.HandleFunc("/play", startPlaySessionHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.Handle("/", http.StripPrefix("", http.FileServer(http.Dir("www"))))
	log.Fatal(http.ListenAndServe(":8888", router))

}

func startPlaySessionHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	var playSession PlaySession
	_ = json.Unmarshal(body, &playSession)

	sessionDuration, err := time.ParseDuration(playSession.Duration)
	if err != nil {
		log.Println("failed to create game session %s", err)
		w.WriteHeader(400)
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	session := gameover.GameSession{
		SessionDuration: sessionDuration,
	}
	gameRequestCh <- gameover.GameRequest{Session: session}
	gameResponse := <-gameResponseCh
	_, _ = fmt.Fprintf(w, "Response: %t %s", gameResponse.Ok, gameResponse.Message)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	gameMaster.Register(gameover.NewWebSocketClient(ws, gameMaster))
}
