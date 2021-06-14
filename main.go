package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// HandleIndex is called when your Battlesnake is created and refreshed
// by play.battlesnake.com. BattlesnakeInfoResponse contains information about
// your Battlesnake, including what it should look like on the game board.
func HandleIndex(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")
	response := snakes[name].Info

	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("ERROR: response write: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Println("Index: " + name)
}

// HandleStart is called at the start of each game your Battlesnake is playing.
// The GameRequest object contains information about the game that's about to start.
// TODO: Use this function to decide how your Battlesnake is going to look on the board.
func HandleStart(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")

	// Nothing to respond with here
	fmt.Println("START: " + name)

	req := GameRequest{}
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	err := json.Unmarshal(b, &req)
	if err != nil {
		fmt.Println("ERROR: start parse: " + err.Error() + ", " + string(b))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fn := snakes[name].Start
	if fn == nil {
		return
	}

	err = fn(r.Context(), req)
	if err != nil {
		fmt.Println("ERROR: start handle: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HandleMove is called for each turn of each game.
// Valid responses are "up", "down", "left", or "right".
// TODO: Use the information in the GameRequest object to determine your next move.
func HandleMove(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	t0 := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), time.Millisecond*500)
	defer cancel()
	name := p.ByName("name")

	req := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("ERROR: move parse: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fn := snakes[name].Move
	if fn == nil {
		return
	}

	m, err := fn(ctx, req)
	if err != nil {
		log.Printf("ERROR: handleMove: %v\n", err)
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	response := MoveResponse{
		Move: m,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Println("ERROR: move response: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var timeout string
	if fmt.Sprint(req.You.Latency) == "0" {
		timeout = " TIMEOUT!"
	}
	log.Printf("Move: %d %v [%vus %sms%s]\n", req.Turn, m, time.Since(t0).Microseconds(), req.You.Latency, timeout)
}

// HandleEnd is called when a game your Battlesnake was playing has ended.
// It's purely for informational purposes, no response required.
func HandleEnd(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	name := p.ByName("name")

	req := GameRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("ERROR: end parse: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var timeout string
	if req.You.Latency == "0" {
		timeout = " TIMEOUT!"
	}
	log.Printf("End %s: %d [%sms%s]\n", name, req.Turn, req.You.Latency, timeout)

	fn := snakes[name].End
	if fn == nil {
		return
	}

	err = fn(r.Context(), req)
	if err != nil {
		fmt.Println("ERROR: end handle: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	bind := os.Getenv("BIND")
	if len(bind) == 0 {
		bind = "localhost:8080"
	}

	router := httprouter.New()
	router.GET("/:name/", HandleIndex)
	router.POST("/:name/start", HandleStart)
	router.POST("/:name/move", HandleMove)
	router.POST("/:name/end", HandleEnd)

	fmt.Printf("Starting Battlesnake Server at http://%s...\n", bind)
	log.Fatal(http.ListenAndServe(bind, router))
}
