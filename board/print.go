package board

import (
	"bytes"

	"github.com/BattlesnakeOfficial/rules"
)

var chars = []rune{'■', '⌀', '●', '⍟', '◘', '☺', '□', '☻'}

func PrintBoard(state *rules.BoardState, outOfBounds []rules.Point) string {

	board := make([][]rune, state.Width)
	for i := range board {
		board[i] = make([]rune, state.Height)
	}
	for y := int32(0); y < state.Height; y++ {
		for x := int32(0); x < state.Width; x++ {
			board[x][y] = '◦'
		}
	}
	for _, oob := range outOfBounds {
		board[oob.X][oob.Y] = '░'
	}
	for _, f := range state.Food {
		board[f.X][f.Y] = '⚕'
	}
	for i, s := range state.Snakes {
		for _, b := range s.Body {
			if b.X < 0 || b.Y < 0 || b.X >= state.Width || b.Y >= state.Height {
				continue
			}
			board[b.X][b.Y] = chars[i]
		}
	}

	var b bytes.Buffer
	for y := state.Height - 1; y >= 0; y-- {
		for x := int32(0); x < state.Width; x++ {
			b.WriteRune(board[x][y])
		}
		b.WriteString("\n")
	}
	return b.String()
}
