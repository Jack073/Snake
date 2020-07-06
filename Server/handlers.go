package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type startReq struct {
	Height int `json:"height"`
	Width  int `json:"width"`
}

type endReq struct {
	Token string `json:"token"`
}

type startOut struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

type moveOutput struct {
	Board  [][]string `json:"board"`
	Alive  bool       `json:"alive"`
	Won    bool       `json:"won"`
	Length int        `json:"length"`
	Eaten  int        `json:"eaten"`
}

type imageReq struct {
	BoardPositions   [][]string `json:"board_positions"`
	HeadColour       [3]int     `json:"head_colour"`
	BodyColour       [3]int     `json:"body_colour"`
	AppleColour      [3]int     `json:"apple_colour"`
	BackGroundColour [3]int     `json:"background_colour"`
	BlockHeight      int        `json:"block_height"`
	BlockWidth       int        `json:"block_width"`
	BorderColour     [3]int     `json:"border_colour"`
	Width            int        `json:"-"`
}

func checkErr(err error) bool {
	return err != nil
}

// StartGame is the handler func for creating a game
func StartGame(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Error When Reading Request Body"}`)
		return
	}

	res := startReq{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Invalid JSON Form"}`)
		return
	}

	if res.Height <= 0 {
		fmt.Fprint(w, `{"error": "Invalid Height parameter passed, must be a positive integer"}`)
		return
	}

	if res.Width <= 0 {
		fmt.Fprint(w, `{"error": "Invalid Width parameter passed, must be a positive integer"}`)
		return
	}

	token := boards.Add(NewBoard(res.Width, res.Height))

	out := startOut{
		Token:   token,
		Message: "Game started successfully, use the provided token for same session requests",
	}

	outStr, err := json.Marshal(out)

	if err != nil {
		fmt.Fprint(w, `{"error": "Error attempting to form output"}`)
		return
	}

	fmt.Fprint(w, string(outStr))

}

// EndGame is the handler func for deleting a game
func EndGame(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Error When Reading Request Body"}`)
		return
	}

	res := endReq{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Invalid JSON Form"}`)
		return
	}

	if res.Token == "" {
		fmt.Fprint(w, `{"error": "No token provided"}`)
		return
	}

	err = boards.Delete(res.Token)

	if err != nil {
		fmt.Fprint(w, `{"error": "Invalid token provided"}`)
		return
	}

	fmt.Fprint(w, `{"message": "Success"}`)

}

// GameMove is the handler func for changing the direction of the snake or moving
func GameMove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Error When Reading Request Body"}`)
		return
	}

	res := struct {
		Token     string `json:"token"`
		Direction string `json:"direction"`
	}{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Invalid JSON Form"}`)
		return
	}

	if res.Token == "" {
		fmt.Fprint(w, `{"error": "Missing Board Token"}`)
		return
	}

	if res.Direction == "" {
		fmt.Fprint(w, `{"error": "Missing direction"}`)
		return
	}

	board := boards.Get(res.Token)

	if board == nil {
		fmt.Fprint(w, `{"error": "Invalid Token"}`)
		return
	}

	direction := strings.ToLower(res.Direction)

	var status int

	switch direction {
	case "u":
		status = board.Snake.Move(0, -1)
	case "d":
		status = board.Snake.Move(0, 1)
	case "r":
		status = board.Snake.Move(1, 0)
	case "l":
		status = board.Snake.Move(-1, 0)
	default:
		fmt.Fprint(w, `{"error": "Invalid Direction"}`)
		return
	}

	boardMap := board.Map()

	if status == 0 {
		out := moveOutput{
			Board:  boardMap,
			Alive:  true,
			Won:    false,
			Length: board.Snake.Length,
			Eaten:  board.Snake.Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		return
	}

	if status == 1 {
		// Collision with self
		out := moveOutput{
			Board:  boardMap,
			Alive:  false,
			Won:    false,
			Length: board.Snake.Length,
			Eaten:  board.Snake.Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = boards.Delete(board.ID)
		// Delete Game, ignoring any errors
		return
	}

	if status == 2 {
		out := moveOutput{
			Board:  boardMap,
			Alive:  true,
			Won:    true,
			Length: board.Snake.Length,
			Eaten:  board.Snake.Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = boards.Delete(board.ID)
		// End game
		return
	}

	if status == 3 {
		// Crashed into wall
		out := moveOutput{
			Board:  boardMap,
			Alive:  false,
			Won:    false,
			Length: board.Snake.Length,
			Eaten:  board.Snake.Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = boards.Delete(board.ID)
		// Crashed into wall
		return
	}
}

func checkColour(c [3]int) bool {
	for _, v := range c {
		if v > 255 {
			return false
		}
	}

	return true
}

// ImageCreate is the handler func for creating images to display
func ImageCreate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Error When Reading Request Body"}`)
		return
	}

	res := imageReq{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"error": "Invalid JSON Form"}`)
		return
	}

	if len(res.BoardPositions) == 0 {
		fmt.Fprint(w, `{"error": "Missing BoardPositions parameter"}`)
		return
	}

	res.Width = len(res.BoardPositions[0])

	for _, r := range res.BoardPositions {
		if len(r) != res.Width {
			fmt.Fprint(w, `{"error": "Inconsistent row width"}`)
			return
		}
	}

	if !checkColour(res.HeadColour) {
		fmt.Fprint(w, `{"error": "Missing HeadColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BodyColour) {
		fmt.Fprint(w, `{"error": "Missing BodyColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.AppleColour) {
		fmt.Fprint(w, `{"error": "Missing AppleColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BorderColour) {
		fmt.Fprint(w, `{"error": "Missing BorderColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BackGroundColour) {
		fmt.Fprint(w, `{"error": "Missing BackGroundColour parameter, maximum value 255"}`)
		return
	}

	if res.BlockHeight <= 0 {
		fmt.Fprint(w, `{"error": "BlockHeight must be greater than 0"}`)
		return
	}

	if res.BlockWidth <= 0 {
		fmt.Fprint(w, `{"error": "BlockWidth must be greater than 0"}`)
		return
	}

	img, err := GenImage(res)

	if err != nil {
		fmt.Fprint(w, `{"error": "Error creating image"}`)
		return
	}

	success := struct {
		Image string `json:"image"`
	}{img}

	out, _ := json.Marshal(success)

	fmt.Fprint(w, string(out))
}
