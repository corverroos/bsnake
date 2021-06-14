package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

type Game struct {
	ID      string `json:"id"`
	Timeout int32  `json:"timeout"`
}

type Coord struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coord) Equal(x Coord) bool {
	return c.X == x.X && c.Y == x.Y
}

func (c Coord) String() string {
	return fmt.Sprintf("[%d:%d]", c.X, c.Y)
}

func (c Coord) Move(m Move) Coord {
	switch m {
	case Up:
		return Coord{c.X, c.Y + 1}
	case Down:
		return Coord{c.X, c.Y - 1}
	case Right:
		return Coord{c.X + 1, c.Y}
	case Left:
		return Coord{c.X - 1, c.Y}
	default:
		panic("unknown move: ")
	}
}

type Battlesnake struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Health int   `json:"health"`
	Body   []Coord `json:"body"`
	Head   Coord   `json:"head"`
	Length int   `json:"length"`
	Shout  string  `json:"shout"`
	Latency interface{}   `json:"latency"`
}

type Board struct {
	Height int           `json:"height"`
	Width  int           `json:"width"`
	Food   []Coord       `json:"food"`
	Snakes []Battlesnake `json:"snakes"`
}

type BattlesnakeInfoResponse struct {
	APIVersion string `json:"apiversion"`
	Author     string `json:"author"`
	Color      string `json:"color"`
	Head       string `json:"head"`
	Tail       string `json:"tail"`
}

type GameRequest struct {
	Game  Game        `json:"game"`
	Turn  int         `json:"turn"`
	Board Board       `json:"board"`
	You   Battlesnake `json:"you"`
}

type MoveResponse struct {
	Move  string `json:"move"`
	Shout string `json:"shout,omitempty"`
}


type Move int

const (
	Up    Move = 1
	Down  Move = 2
	Right Move = 3
	Left  Move = 4
)


func (m Move) String() string {
	switch m {
	case Up:
		return "up"
	case Down:
		return "down"
	case Right:
		return "right"
	case Left:
		return "left"
	default:
		panic("unknown move: ")
	}
}

var Moves = []Move{Up, Down, Right, Left}

func RandMoves() []Move {
	var res []Move
	for _, move := range Moves {
		res = append(res, move)
	}
	rand.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}


type Area map[Coord]int

func (a Area) Size() int {
	var res int
	for _, v := range a {
		if v > 0 {
			res++
		}
	}
	return res
}

func (a Area) MinTTL() (Coord, int) {
	min := math.MaxInt64
	var c Coord
	var found bool
	for coord, v := range a {
		if v >= 0 {
			continue
		}
		if -v < min {
			c = coord
			min = -v
			found = true
		}
	}
	if ! found {
		return Coord{}, 0
	}
	return c, min
}

func (a Area) Viz() string {
	var res [][]rune
	const (
		s  = '.'
		f = '*'
	)
	var maxY, maxX int
	for c, b := range a {
		if b == 0 {
			continue
		}
		if maxX < c.X {
			maxX = c.X
		}
		if maxY < c.Y {
			maxY = c.Y
		}
	}


	for y := 0; y <= maxY; y++ {
		var row []rune
		for x := 0; x <= maxX; x++ {
			row = append(row, s)
		}
		res = append(res, row)
	}

	for c, b := range a {
		if b == 0 {
			continue
		}

		if b > 0 {
			res[c.Y][c.X] = f
		} else if b <= -10 {
			res[c.Y][c.X] = 'x'
		} else {
			res[c.Y][c.X] = []rune(fmt.Sprint(-b))[0]
		}
	}
	var sl []string
	for _, row := range res {
		sl = append([]string{string(row)}, sl...)
	}
	return strings.Join(sl, "\n")
}
