package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type startReq struct {
	Height, Width int
}

type endReq struct {
	Token string
}

type startOut struct {
	Token   string
	Message string
}

type moveOutput struct {
	Board  [][]string
	Alive  bool
	Won    bool
	Length int
	Eaten  int
}

type ImageReq struct {
	BoardPositions   [][]string
	HeadColour       [3]int
	BodyColour       [3]int
	AppleColour      [3]int
	BackGroundColour [3]int
	BlockHeight      int
	BlockWidth       int
	BorderColour     [3]int
	width            int
	// Unexported field can't be filled by json.Unmarshal
}

func checkErr(err error) bool {
	if err == nil {
		return false
	}

	return true
}

func StartGame(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Error When Reading Request Body"}`)
		return
	}

	res := startReq{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Invalid JSON Form"}`)
		return
	}

	if res.Height <= 0 {
		fmt.Fprint(w, `{"Error": "Invalid Height parameter passed, must be a positive integer"}`)
		return
	}

	if res.Width <= 0 {
		fmt.Fprint(w, `{"Error": "Invalid Width parameter passed, must be a positive integer"}`)
		return
	}

	token := Boards.Add(NewBoard(res.Width, res.Height))

    fmt.Println("Created game with token:", token)

	out := startOut{
		Token:   token,
		Message: "Game started successfully, use the provided token for same session requests",
	}

	outStr, err := json.Marshal(out)

	if err != nil {
		fmt.Fprint(w, `{"Error": "Error attempting to form output"}`)
		return
	}

	fmt.Fprint(w, string(outStr))

}

func EndGame(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Error When Reading Request Body"}`)
		return
	}

	res := endReq{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Invalid JSON Form"}`)
		return
	}

	if res.Token == "" {
		fmt.Fprint(w, `{"Error": "No token provided"}`)
		return
	}

	err = Boards.Delete(res.Token)

	if err != nil {
		fmt.Fprint(w, `{"Error": "Invalid token provided"}`)
		return
	}

	fmt.Fprint(w, `{"Message": "Success"}`)

}

func GameMove(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Error When Reading Request Body"}`)
		return
	}

	res := struct {
		Token     string
		Direction string
	}{}

	err = json.Unmarshal(body, &res)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Invalid JSON Form"}`)
		return
	}

	if res.Token == "" {
		fmt.Fprint(w, `{"Error": "Missing Board Token"}`)
		return
	}

	if res.Direction == "" {
		fmt.Fprint(w, `{"Error": "Missing direction"}`)
		return
	}

	board := Boards.Get(res.Token)

	if board == nil {
		fmt.Fprint(w, `{"Error": "Invalid Token"}`)
		return
	}

    fmt.Println("Moved snake in game", res.Token, "direction", res.Direction)

	direction := strings.ToLower(res.Direction)

	var status int

	if direction == "u" {
		status = (*(*board).Snake).Move(0, -1)
	} else if direction == "d" {
		status = (*(*board).Snake).Move(0, 1)
	} else if direction == "r" {
		status = (*(*board).Snake).Move(1, 0)
	} else if direction == "l" {
		status = (*(*board).Snake).Move(-1, 0)
	} else {
		fmt.Fprint(w, `{"Error": "Invalid Direction"}`)
		return
	}

	boardMap := (*board).Map()

	if status == 0 {
		out := moveOutput{
			Board:  boardMap,
			Alive:  true,
			Won:    false,
			Length: (*(*board).Snake).Length,
			Eaten:  (*(*board).Snake).Eaten,
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
			Length: (*(*board).Snake).Length,
			Eaten:  (*(*board).Snake).Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = Boards.Delete((*board).ID)
		// Delete Game, ignoring any errors
		return
	}

	if status == 2 {
		out := moveOutput{
			Board:  boardMap,
			Alive:  true,
			Won:    true,
			Length: (*(*board).Snake).Length,
			Eaten:  (*(*board).Snake).Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = Boards.Delete((*board).ID)
		// End game
		return
	}

	if status == 3 {
		// Crashed into wall
		out := moveOutput{
			Board:  boardMap,
			Alive:  false,
			Won:    false,
			Length: (*(*board).Snake).Length,
			Eaten:  (*(*board).Snake).Eaten,
		}

		outStr, _ := json.Marshal(out)

		fmt.Fprint(w, string(outStr))

		_ = Boards.Delete((*board).ID)
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

func ImageCreate(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if checkErr(err) {
		fmt.Fprint(w, `{"Error": "Error When Reading Request Body"}`)
		return
	}

	res := ImageReq{}

	err = json.Unmarshal(body, &res)

	if len(res.BoardPositions) == 0 {
		fmt.Fprint(w, `{"Error": "Missing BoardPositions parameter"}`)
		return
	}

	res.width = len(res.BoardPositions[0])

	for _, r := range res.BoardPositions {
		if len(r) != res.width {
			fmt.Fprint(w, `{"Error": "Inconsistent row width"}`)
			return
		}
	}

	if !checkColour(res.HeadColour) {
		fmt.Fprint(w, `{"Error": "Missing HeadColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BodyColour) {
		fmt.Fprint(w, `{"Error": "Missing BodyColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.AppleColour) {
		fmt.Fprint(w, `{"Error": "Missing AppleColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BorderColour) {
		fmt.Fprint(w, `{"Error": "Missing BorderColour parameter, maximum value 255"}`)
		return
	}

	if !checkColour(res.BackGroundColour) {
		fmt.Fprint(w, `{"Error": "Missing BackGroundColour parameter, maximum value 255"}`)
		return
	}

	if res.BlockHeight <= 0 {
		fmt.Fprint(w, `{"Error": "BlockHeight must be greater than 0"}`)
		return
	}

	if res.BlockWidth <= 0 {
		fmt.Fprint(w, `{"Error": "BlockWidth must be greater than 0"}`)
		return
	}

	img, err := GenImage(res)

	if err != nil {
		fmt.Fprint(w, `{"Error": "Error creating image"}`)
		return
	}

	success := struct {
		Image string
	}{img}

	out, _ := json.Marshal(success)

	fmt.Fprint(w, string(out))
}
