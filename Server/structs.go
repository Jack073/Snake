package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"math/rand"
	"strconv"
	"time"
)

type Board struct {
	ID     string // Some ID to represent the game code
	Height int
	Width  int
	Snake  *Snake
	Apple  *Apple
}

func (r *RunningGames) Exists(game string) bool {
	return (*r).find(game) != -1
}

func (r *RunningGames) find(game string) int {
	for l, g := range *r {
		if g.ID == game {
			return l
		}
	}

	return -1
}

func (r *RunningGames) Delete(game string) error {
	if !(*r).Exists(game) {
		return errors.New("Invalid Token")
	}

	loc := (*r).find(game)

	if loc == -1 {
		// How did you do this
		return errors.New("Invalid Token")
	}

	*r = append((*r)[:loc], (*r)[loc+1:]...)

	return nil
}

type RunningGames []*Board

func (r *RunningGames) Add(n *Board) string {
	*r = append(*r, n)

	return n.ID
}

func (r *RunningGames) Get(gameId string) *Board {
	for _, board := range *r {
		if board.ID == gameId {
			return board
		}
	}

	return nil
}

func NewBoard(Width, Height int) *Board {
	b := &Board{
		ID:     NewID(),
		Width:  Width,
		Height: Height,
	}

	b.Snake = NewSnake(b)

	(*b).Apple = &Apple{Board: b}

	appleCoords, err := (*b).Apple.GenApple()

	if err != nil {
		appleCoords = [2]int{10, 10}
	}

	(*b).Apple.Coords = appleCoords

	return b
}

type Snake struct {
	Head   *Block
	Length int
	Eaten  int
	Board  *Board
}

func (s *Snake) String() string {
	st := ""

	h := (*s).Head
	for (*h).Next != nil {
		st += (*h).Format()

		h = (*h).Next

		if (*h).Next == nil {
			st += (*h).Format()
		}
	}

	return st
}

type Block struct {
	Position [2]int
	Next     *Block
	Previous *Block
}

func (b *Block) Format() string {
	return formatCoords((*b).Position)
}

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

func (s *Snake) Add(n int) {
	start := *(*s).Head

	connectedPointer := start.Next

	for start.Next != nil {
		connectedPointer = start.Next
		start = *connectedPointer
	}

	for i := 0; i < n; i++ {
		start = *connectedPointer
		var newPos [2]int

		if (*start.Previous).Position[0] == start.Position[0]+1 {
			// Going Right
			// Spawn Left
			newPos = [2]int{
				start.Position[0] - 1,
				start.Position[1],
			}
		} else if (*start.Previous).Position[0] == start.Position[0]-1 {
			// Going Left
			// Spawn Right
			newPos = [2]int{
				start.Position[0] + 1,
				start.Position[1],
			}
		} else if (*start.Previous).Position[1] == start.Position[1]+1 {
			// Going Down
			// Spawn Up
			newPos = [2]int{
				start.Position[0],
				start.Position[1] - 1,
			}
		} else if (*start.Previous).Position[1] == start.Position[1]-1 {
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

		if newPos[0] >= (*(*s).Board).Width {
			newPos[0] = (*(*s).Board).Width
		}

		if newPos[1] >= (*(*s).Board).Height {
			newPos[1] = (*(*s).Board).Height
		}

		newBlock := &Block{
			Position: newPos,
			Previous: connectedPointer,
		}

		(*connectedPointer).Next = newBlock

		connectedPointer = newBlock

	}
}

func NewID() string {
	hasher := md5.New()

	timeNow := time.Now().Format("2006-01-02T15:04:05-0700")

	timeNow += strconv.FormatInt(time.Now().UnixNano(), 10)

	hasher.Write([]byte(timeNow))

	for res := hex.EncodeToString(hasher.Sum(nil)); Boards.Get(res) != nil; {
		hasher.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	}

	return hex.EncodeToString(hasher.Sum(nil))
}

type Apple struct {
	Coords [2]int
	Board  *Board
}

func sliceMatch(slice1, slice2 [2]int) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	for l, v := range slice1 {
		if v != slice2[l] {
			return false
		}
	}

	return true
}

func sortIntSlice(slice [][2]int) [][2]int {
	for {
		changed := false
		for n := 0; n < len(slice)-1; n++ {
			if slice[n][0] > slice[n+1][0] {
				changed = true
				temp := slice[n]
				slice[n] = slice[n+1]
				slice[n+1] = temp
			}

			if slice[n][0] == slice[n+1][0] {
				if slice[n][1] > slice[n+1][1] {
					changed = true
					temp := slice[n]
					slice[n] = slice[n+1]
					slice[n+1] = temp
				}
			}
		}

		if !changed {
			return slice
		}
	}
}

func divUp(a, b int) int {
	// Function which returns the result of a / b, rounded up to a whole number

	if a%b == 0 {
		return a / b
	}

	return 1 + (a / b)
	// return a / b
}

func (a *Apple) GenApple() ([2]int, error) {
	board := *((*a).Board)

	snakePositions := make([][2]int, 0)

	snake := *(board.Snake)

	nextSnake := snake.Head

	for (*nextSnake).Next != nil {
		snakePositions = append(snakePositions, (*nextSnake).Position)

		nextSnake = (*nextSnake).Next

		if (*nextSnake).Next == nil {
			snakePositions = append(snakePositions, (*nextSnake).Position)
		}
	}

	if len(snakePositions) == (board.Width * board.Height) {
		return [2]int{-1, -1}, errors.New("No blank spaces")
	}

	totalRandom := 0

	final := [2]int{-1, -1}

	snakePositions = sortIntSlice(snakePositions)

	for {
		totalRandom++

		x := rand.Intn(board.Width)

		y := rand.Intn(board.Height)

		coord := [2]int{x, y}

		hit := false

		pos := divUp(len(snakePositions), 2)

		for len(snakePositions) > 0 && !hit {

			pos = divUp(len(snakePositions), 2)

			if len(snakePositions) == 1 {
				hit = sliceMatch(snakePositions[0], coord)
				break
			}

			if sliceMatch(snakePositions[pos], coord) {
				hit = true
				break
			}

			if len(snakePositions) < 2 {
				break
			}

			/* if snakePositions[pos][0] > coord[0] {
				snakePositions = snakePositions[pos+1:]
			} else if snakePositions[pos][0] < coord[0] {
				snakePositions = snakePositions[:pos]
			} else if snakePositions[pos][0] == coord[0] {
				if snakePositions[pos][1] == coord[1] {
					hit = true
					break
				} else if snakePositions[pos][1] < coord[1] {
					snakePositions = snakePositions[:pos]
				} else {
					snakePositions = snakePositions[pos+1:]
				}
			}*/
			if snakePositions[pos][0] > coord[0] {
                snakePositions = snakePositions[:pos]
            } else if snakePositions[pos][0] < coord[0] {
            	snakePositions = snakePositions[pos+1:]
            } else if snakePositions[pos][0] == coord[0] {
            	if snakePositions[pos][1] == coord[1] {
            		hit = true
            		break
            	} else if snakePositions[pos][1] < coord[1] {
            		snakePositions = snakePositions[pos+1:]
            	} else {
            		snakePositions = snakePositions[:pos]
            	}
            }
		}

		if totalRandom >= 30 && hit {
			// Random is taking too long, iterate over the grid and select the first empty grid

			for x = 0; x < board.Width; x++ {
				for y = 0; y < board.Height; y++ {
					hit = false

					cGen := [2]int{x, y}

					if snakePositions[pos][0] > coord[0] {
						snakePositions = snakePositions[pos+1:]
					} else if snakePositions[pos][0] < coord[0] {
						snakePositions = snakePositions[:pos]
					} else if snakePositions[pos][0] == coord[0] {
						if snakePositions[pos][1] == coord[1] {
							hit = true
							break
						} else if snakePositions[pos][1] < coord[1] {
							snakePositions = snakePositions[:pos]
						} else {
							snakePositions = snakePositions[pos+1:]
						}
					}

					if !hit {
						return cGen, nil
					}
				}
			}

			// How did you do this?
			return final, errors.New("No empty spaces")
		}

		if hit {
			continue
		}

		final = coord
		break
	}

	return final, nil
}

func formatCoords(coords [2]int) string {
	return strconv.Itoa(coords[0]) + "." + strconv.Itoa(coords[1])
}

func (b *Board) Map() [][]string {
	var rep string

	snakePositions := make(map[string]string)

	head := (*(*b).Snake).Head

	for (*head).Next != nil {
		if (*head).Previous == nil {
			rep = "h"
		} else {
			rep = "s"
		}

		snakePositions[(*head).Format()] = rep

		head = (*head).Next

		if (*head).Next == nil {
			if (*head).Previous == nil {
				rep = "h"
			} else {
				rep = "s"
			}

			snakePositions[(*head).Format()] = rep
		}
	}

	apple := (*b).Apple

	rows := make([][]string, (*b).Height)

	for y := 0; y < (*b).Height; y++ {
		row := make([]string, (*b).Width)
		for x := 0; x < (*b).Width; x++ {
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

func (s *Snake) Move(x, y int) int {
	head := (*s).Head

	oldHeadPosition := (*head).Position

	(*head).Position[0] += x
	(*head).Position[1] += y

	newHeadPosition := (*head).Position

	start := head

	if (*head).Position[0] == (*(*s).Board).Apple.Coords[0] && (*head).Position[1] == (*(*s).Board).Apple.Coords[1] {
		(*s).Add(1)
		(*s).Length++
		(*s).Eaten++
		newApple, err := (*(*s).Board).Apple.GenApple()

		if err != nil {
			return 2
			// No where else to place the apple
		}

		(*(*(*s).Board).Apple).Coords = newApple
	}

	var coords [2]int
	coordsNew := oldHeadPosition

	head = (*head).Next

	for (*head).Next != nil {
		if (*head).Position[0] == newHeadPosition[0] && (*head).Position[1] == newHeadPosition[1] {
			return 1
			// Collision, end game
		}

		coords = (*head).Position
		(*head).Position = coordsNew
		coordsNew = coords
		head = head.Next

		if (*head).Next == nil {
			// Last one
			(*head).Position = coordsNew
		}
	}

	if (*start).Position[0] < 0 || (*start).Position[0] >= (*(*s).Board).Width {
		return 3

		// Crash into wall (width)
	} else if (*start).Position[1] < 0 || (*start).Position[1] >= (*(*s).Board).Height {
		return 3

		// Crash into wall (Height)
	}

	return 0
	// Nothing with noting happened
}
