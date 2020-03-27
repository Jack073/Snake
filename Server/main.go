package main

import (
	"fmt"
	"log"
	"net/http"
)

const VERSION = 1.0

var (
	Boards = make(RunningGames, 0)
)

func main() {
	http.HandleFunc("/start", StartGame)

	http.HandleFunc("/destroy", EndGame)

	http.HandleFunc("/move", GameMove)

	http.HandleFunc("/image", ImageCreate)

	fmt.Println("Snake Server Version ", VERSION, " Now Online, Listening For Requests At Port 8081")

	log.Fatal(http.ListenAndServe(":8081", nil))
}
