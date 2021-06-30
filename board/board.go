package board

import (
	"math/rand"

	"github.com/BattlesnakeOfficial/rules"
)

var Moves = []string{"up", "down", "right", "left"}

func RandMoves() []string {
	return moveperms[rand.Intn(perms)]
}

func GenMoveSet(board *rules.BoardState) [][]string {
	res := [][]string{make([]string, len(board.Snakes))}

	clone := func(m []string) []string {
		return append([]string(nil), m...)
	}

	for i := 0; i < len(board.Snakes); i++ {
		if board.Snakes[i].EliminatedCause != "" {
			continue
		}

		temp := make([][]string, 0, 4*len(res))
		for mi, move := range Moves {
			if !IsRationalMove(board, i, move) {
				// Skip unless it will result in 0 moves
				if len(temp) > 0 || mi < 3 {
					continue
				}
			}

			for _, prev := range res {
				next := clone(prev)
				next[i] = move
				temp = append(temp, next)
			}
		}

		res = temp
	}
	return res
}

func IsRationalMove(board *rules.BoardState, snakeIdx int, move string) bool {
	next := MovePoint(board.Snakes[snakeIdx].Body[0], move)

	if next.X < 0 || next.X >= board.Width {
		return false
	}

	if next.Y < 0 || next.Y >= board.Height {
		return false
	}

	for i := 0; i < len(board.Snakes); i++ {
		for j := 0; j < len(board.Snakes[i].Body)-1; j++ {
			if next.X == board.Snakes[i].Body[j].X && next.Y == board.Snakes[i].Body[j].Y {
				return false
			}
		}
	}

	return true
}

func IsLoosingH2H(board *rules.BoardState, snakeIdx int, move string) bool {
	next := MovePoint(board.Snakes[snakeIdx].Body[0], move)

	for i := 0; i < len(board.Snakes); i++ {
		if i == snakeIdx || len(board.Snakes[i].Body) <= len(board.Snakes[snakeIdx].Body) {
			continue
		}
		if Distance(next, board.Snakes[i].Body[0]) == 1 {
			return true
		}
	}

	return false
}

func MovePoint(p rules.Point, move string) rules.Point {
	switch move {
	case "up":
		return rules.Point{X: p.X, Y: p.Y + 1}
	case "down":
		return rules.Point{X: p.X, Y: p.Y - 1}
	case "left":
		return rules.Point{X: p.X - 1, Y: p.Y}
	case "right":
		return rules.Point{X: p.X + 1, Y: p.Y}
	}
	panic("unknown move")
}

func Distance(a, b rules.Point) int32 {
	x := a.X - b.X
	if x < 0 {
		x = -x
	}
	y := a.Y - b.Y
	if y < 0 {
		y = -y
	}

	return x + y
}
