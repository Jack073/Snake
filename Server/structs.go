package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// Board contains the information for a single game
type Board struct {
	ID     string // Some ID to represent the game code
	Height int
	Width  int
	Snake  *Snake
	Apple  *Apple
	sync.RWMutex
}

// Exists returns whether or not a game with the given token exists
func (r *RunningGames) Exists(game string) bool {
	return r.find(game) > -1
}

// The location of the game with the given token (given it exists, else -1)
func (r *RunningGames) find(game string) int {
	r.Lock()

	defer r.Unlock()

	for l, g := range r.boards {
		if g.ID == game {
			return l
		}
	}

	return -1
}

// Delete stops (deletes) a game with the given token
func (r *RunningGames) Delete(game string) error {
	if !r.Exists(game) {
		return errors.New("Invalid Token")
	}

	loc := r.find(game)
	if loc == -1 {
		// How did you do this
		return errors.New("Invalid Token")
	}
	r.Lock()
	if len(r.boards) == 1 {
		// Reset to empty slice
		r.boards = make([]*Board, 0)
	} else {
		r.boards[loc] = r.boards[len(r.boards)-1]
		r.boards = r.boards[:len(r.boards)-1]
	}
	r.Unlock()

	return nil
}

// RunningGames is a struct containing all active game boards
type RunningGames struct {
	sync.RWMutex
	boards []*Board
}

// Add appends the game to the active games
func (r *RunningGames) Add(n *Board) string {
	r.Lock()
	r.boards = append(r.boards, n)
	r.Unlock()

	return n.ID
}

// Get returns the board with the given ID (should it exist, else nil)
func (r *RunningGames) Get(gameID string) *Board {
	r.Lock()

	defer r.Unlock()

	for _, board := range r.boards {
		if board.ID == gameID {
			return board
		}
	}

	return nil
}

// NewBoard returns a new Board struct with all required fields filled
func NewBoard(Width, Height int) *Board {
	b := &Board{
		ID:     NewID(),
		Width:  Width,
		Height: Height,
	}

	b.Snake = NewSnake(b)

	b.Apple = &Apple{Board: b}

	appleCoords, err := b.Apple.GenApple()

	if err != nil {
		appleCoords = [2]int{10, 10}
	}

	b.Apple.Coords = appleCoords

	return b
}

// Snake represents a snake in the game, containing all required info
type Snake struct {
	Head   *Block
	Length int
	Eaten  int
	Board  *Board
	sync.RWMutex
}

// Block represents a single block in the game
type Block struct {
	Position [2]int
	Next     *Block
	Previous *Block
}

// Format is used to help create a "map" of the game
func (b *Block) Format() string {
	return formatCoords(b.Position)
}

// NewSnake returns a new snake, initialising the length
func NewSnake(b *Board) *Snake {
	head := &Block{
		Position: [2]int{0, 5},
		Next:     nil,
		Previous: nil,
	}

	nextBlock := &Block{
		Position: [2]int{0, 4},
		Next:     nil,
		Previous: head,
	}

	head.Next = nextBlock

	s := Snake{
		Head:   head,
		Length: 2,
		Eaten:  0,
		Board:  b,
	}

	s.Add(1)

	s.Length++

	return &s
}

// Add makes the snake add n blocks to its tail, growing it
func (s *Snake) Add(n int) {
	// No Mutex lock as this should only be called by locked methods anyway
	start := s.Head

	connectedPointer := start.Next

	for start.Next != nil {
		connectedPointer = start.Next
		start = connectedPointer
	}

	for i := 0; i < n; i++ {
		start = connectedPointer
		var newPos [2]int

		if start.Previous.Position[0] == start.Position[0]+1 {
			// Going Right
			// Spawn Left
			newPos = [2]int{
				start.Position[0] - 1,
				start.Position[1],
			}
		} else if start.Previous.Position[0] == start.Position[0]-1 {
			// Going Left
			// Spawn Right
			newPos = [2]int{
				start.Position[0] + 1,
				start.Position[1],
			}
		} else if start.Previous.Position[1] == start.Position[1]+1 {
			// Going Down
			// Spawn Up
			newPos = [2]int{
				start.Position[0],
				start.Position[1] - 1,
			}
		} else if start.Previous.Position[1] == start.Position[1]-1 {
			// Going Up
			// Spawn Down
			newPos = [2]int{
				start.Position[0],
				start.Position[1] + 1,
			}
		}

		if newPos[0] < 0 {
			newPos[0] = 0
		}

		if newPos[1] < 0 {
			newPos[1] = 0
		}

		if newPos[0] >= s.Board.Width {
			newPos[0] = s.Board.Width
		}

		if newPos[1] >= s.Board.Height {
			newPos[1] = s.Board.Height
		}

		newBlock := &Block{
			Position: newPos,
			Previous: connectedPointer,
		}

		connectedPointer.Next = newBlock

		connectedPointer = newBlock

	}
}

// NewID generates a game token, (using a md5 hash for simplicity) and ensures its unique
func NewID() string {
	hasher := md5.New()

	timeNow := time.Now().Format("2006-01-02T15:04:05-0700")

	timeNow += strconv.FormatInt(time.Now().UnixNano(), 10)

	_, _ = hasher.Write([]byte(timeNow))

	for res := hex.EncodeToString(hasher.Sum(nil)); boards.Get(res) != nil; {
		_, _ = hasher.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

// Apple represents the apple on the board
type Apple struct {
	Coords [2]int
	Board  *Board
	sync.RWMutex
}

// GenApple returns new coords for an apple, or an error if unable to
func (a *Apple) GenApple() ([2]int, error) {
	a.Lock()

	defer a.Unlock()

	board := a.Board

	snakePositions := make(map[[2]int]interface{})

	snake := board.Snake

	nextSnake := snake.Head

	for nextSnake.Next != nil {
		snakePositions[nextSnake.Position] = nil

		nextSnake = nextSnake.Next

		if nextSnake.Next == nil {
			snakePositions[nextSnake.Position] = nil
		}
	}

	if len(snakePositions) == (board.Width * board.Height) {
		return [2]int{-1, -1}, errors.New("No blank spaces")
	}

	passed := 0

	location := rand.Intn(a.Board.Width*a.Board.Height - len(snakePositions))

	totalPassed := 0

	var coords [2]int

	for passed <= location {
		coords = [2]int{
			totalPassed % a.Board.Width,
			totalPassed / a.Board.Width,
		}

		if _, ok := snakePositions[coords]; ok {
			passed--
		}
		passed++
		totalPassed++
	}

	return coords, nil
}

func formatCoords(coords [2]int) string {
	return strconv.Itoa(coords[0]) + "." + strconv.Itoa(coords[1])
}

// Map creates a "map" of the board to return to the user
func (b *Board) Map() [][]string {
	b.Lock()

	defer b.Unlock()

	var rep string

	snakePositions := make(map[string]string)

	head := b.Snake.Head

	for head.Next != nil {
		if head.Previous == nil {
			rep = "h"
		} else {
			rep = "s"
		}

		snakePositions[head.Format()] = rep

		head = head.Next

		if head.Next == nil {
			if head.Previous == nil {
				rep = "h"
			} else {
				rep = "s"
			}

			snakePositions[head.Format()] = rep
		}
	}

	apple := b.Apple

	rows := make([][]string, b.Height)

	for y := 0; y < b.Height; y++ {
		row := make([]string, b.Width)
		for x := 0; x < b.Width; x++ {
			if apple.Coords[0] == x && apple.Coords[1] == y {
				row[x] = "a"
			} else if p, ok := snakePositions[formatCoords([2]int{x, y})]; ok {
				row[x] = p
			} else {
				row[x] = " "
			}
		}
		rows[y] = row
	}

	return rows
}

// Move moves the snake by the given x and y vector
func (s *Snake) Move(x, y int) int {
	s.Lock()

	defer s.Unlock()

	head := s.Head

	oldHeadPosition := head.Position

	head.Position[0] += x
	head.Position[1] += y

	newHeadPosition := head.Position

	start := head

	if head.Position[0] == s.Board.Apple.Coords[0] && head.Position[1] == s.Board.Apple.Coords[1] {
		s.Add(1)
		s.Length++
		s.Eaten++
		newApple, err := s.Board.Apple.GenApple()

		if err != nil {
			return 2
			// No where else to place the apple
		}

		s.Board.Apple.Coords = newApple
	}

	var coords [2]int
	coordsNew := oldHeadPosition

	head = head.Next

	for head.Next != nil {
		if head.Position[0] == newHeadPosition[0] && head.Position[1] == newHeadPosition[1] {
			return 1
			// Collision, end game
		}

		coords = head.Position
		head.Position = coordsNew
		coordsNew = coords
		head = head.Next

		if head.Next == nil {
			// Last one
			head.Position = coordsNew
		}
	}

	if start.Position[0] < 0 || start.Position[0] >= s.Board.Width {
		return 3
		// Crash into wall (width)
	} else if start.Position[1] < 0 || start.Position[1] >= s.Board.Height {
		return 3
		// Crash into wall (Height)
	}

	return 0
	// Nothing worth noting happened
}
