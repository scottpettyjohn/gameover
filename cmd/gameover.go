package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/scottpettyjohn/gameover"
	"github.com/stianeikeland/go-rpio"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type PlaySession struct {
	Duration string `json:"duration"`
}

func (ps PlaySession) parseDuration() time.Duration {
	return 0
}

var (
	gameRequestCh  chan gameover.GameRequest  = make(chan gameover.GameRequest)
	gameResponseCh chan gameover.GameResponse = make(chan gameover.GameResponse)
)

func main() {

	gOver := gameover.Init(gameRequestCh, gameResponseCh)
	gOver.Register(&gameover.Foley{})
	log.Println("ready player 1.")
	router := mux.NewRouter()
	router.HandleFunc("/play", StartPlaySession).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))

}

func StartPlaySession(w http.ResponseWriter, r *http.Request) {
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

func startPlaySession(playSession PlaySession) {
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer rpio.Close()
}
