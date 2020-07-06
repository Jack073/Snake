package main

import (
	"fmt"
	"log"
	"net/http"
)

const version = 1.1

var (
	boards = RunningGames{boards: make([]*Board, 0, 5)}
)

func main() {
	http.HandleFunc("/start", StartGame)

	http.HandleFunc("/destroy", EndGame)

	http.HandleFunc("/move", GameMove)

	http.HandleFunc("/image", ImageCreate)

	fmt.Println("Snake Server Version ", version, " Now Online, Listening For Requests At Port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
