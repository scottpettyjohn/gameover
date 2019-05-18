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
	gameRequestCh  = make(chan gameover.GameRequest)
	gameResponseCh = make(chan gameover.GameResponse)
	upgrader       = websocket.Upgrader{
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
	router.HandleFunc("/extend", extendPlaySessionHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/dist/ui/")))
	s := http.StripPrefix("/home", http.FileServer(http.Dir("./ui/dist/ui/")))
	router.PathPrefix("/home").Handler(s)

	log.Fatal(http.ListenAndServe(":8888", router))
}

func startPlaySessionHandler(w http.ResponseWriter, r *http.Request) {
	playSession := parsePlaySession(r)

	sessionDuration, err := time.ParseDuration(playSession.Duration)
	if err != nil {
		log.Printf("failed to create game session %s", err)
		w.WriteHeader(400)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}
	session := gameover.GameSession{
		SessionDuration: sessionDuration,
	}
	gameRequestCh <- gameover.GameRequest{Session: session}
	gameResponse := <-gameResponseCh
	_, _ = fmt.Fprintf(w, "Response: %t %s", gameResponse.Ok, gameResponse.Message)
}

func extendPlaySessionHandler(w http.ResponseWriter, r *http.Request) {
	playSession := parsePlaySession(r)
	extendDuration, err := time.ParseDuration(playSession.Duration)
	if err != nil {
		log.Printf("failed to extend game session %s", err)
		w.WriteHeader(400)
		_, _ = fmt.Fprintf(w, "error: %s", err)
		return
	}
	extendSession := gameover.GameSession{ExtendSession: true, SessionDuration: extendDuration}
	gameRequestCh <- gameover.GameRequest{Session: extendSession}
	gameResponse := <-gameResponseCh
	_, _ = fmt.Fprintf(w, "Response: %t %s", gameResponse.Ok, gameResponse.Message)
}

func parsePlaySession(r *http.Request) PlaySession {
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(string(body))
	var playSession PlaySession
	_ = json.Unmarshal(body, &playSession)
	return playSession
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	gameMaster.Register(gameover.NewWebSocketClient(ws, gameMaster))
}
