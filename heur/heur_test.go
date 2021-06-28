package heur

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"
)

func TestLength(t *testing.T) {
	tests := []struct {
		Exp  map[int]float64
		Lens []int
	}{
		{
			Exp: map[int]float64{},
		},
		{
			Lens: []int{3, 3, 3},
			Exp:  map[int]float64{0: 0, 1: 0, 2: 0},
		},
		{
			Lens: []int{33, 13, 03},
			Exp:  map[int]float64{0: 0.34013605442176875, 1: -0.06802721088435372, 2: -0.27210884353741494},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			b := rules.BoardState{}
			for _, l := range test.Lens {
				var bl []rules.Point
				for i := 0; i < l; i++ {
					bl = append(bl, rules.Point{})
				}
				b.Snakes = append(b.Snakes, rules.Snake{Body: bl})
			}

			res := Length(&b)
			normalize(res)
			require.EqualValues(t,
				test.Exp,
				res)
		})
	}
}

func TestFlood(t *testing.T) {
	tests := []struct {
		Name     string
		ControlA map[int]float64
		ControlB map[int]float64
		Starve   map[int]bool
		Heur     map[int]float64
		Move     string
	}{
		{
			Name:     "../testdata/input-001.json",
			ControlB: map[int]float64{0: 49},
			Starve:   map[int]bool{0: false},
			Heur:     map[int]float64{0: 0},
			Move:     "down",
		},
		{
			Name:     "../testdata/input-006.json",
			ControlB: map[int]float64{0: 1, 1: 120},
			Starve:   map[int]bool{1: false},
			Heur:     map[int]float64{0: -3.5697887970615247, 1: 3.5697887970615247},
			Move:     "up",
		},
		{
			Name:     "../testdata/input-007.json",
			ControlB: map[int]float64{0: 91, 1: 14, 2: 16},
			Starve:   map[int]bool{0: false, 2: false},
			Heur:     map[int]float64{0: 3.760330578512397, 1: -1.2964876033057848, 2: -2.4638429752066116},
			Move:     "up",
		}, {
			Name:     "../testdata/input-016.json",
			ControlB: map[int]float64{0: 8, 1: 1},
			Starve:   map[int]bool{},
			Heur:     map[int]float64{0: 3.1944444444444446, 1: -3.1944444444444446},
			Move:     "up",
		}, {
			Name:     "../testdata/input-017.json",
			ControlB: map[int]float64{0: 1, 1: 8},
			Starve:   map[int]bool{},
			Heur:     map[int]float64{0: -3.1944444444444446, 1: 3.1944444444444446},
			Move:     "left",
		}, {
			Name:     "../testdata/input-022.json",
			ControlB: map[int]float64{0: 27, 1: 40, 2: 54},
			Starve:   map[int]bool{1: false, 2: false},
			Heur:     map[int]float64{0: -0.780849244799088, 1: 0.7908235964662298, 2: -0.009974351667141557},
			Move:     "up",
		}, {
			Name:     "../testdata/input-029.json",
			ControlB: map[int]float64{0: 3, 1: 22},
			Starve:   map[int]bool{},
			Heur:     map[int]float64{0: -3.0538461538461537, 1: 3.0538461538461537},
			Move:     "left",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, youIdx := fileToBoard(t, test.Name)
			fmt.Printf("YouIdx: %d\n", youIdx)

			control, starve := Flood(b, nil)
			require.EqualValues(t, test.ControlB, control)
			require.EqualValues(t, test.Starve, starve)

			f := Factors{
				Control: 5,
				Length:  10,
				Starve:  -10,
			}
			heur := Calc(f, b, nil)
			require.EqualValues(t, test.Heur, heur)

			move, _ := SelectMove(f, b, nil, youIdx)
			require.Equal(t, test.Move, move)
		})
	}
}

func fileToBoard(t *testing.T, file string) (*rules.BoardState, int) {
	f, err := os.Open(file)
	jtest.RequireNil(t, err)
	var req struct {
		Board rules.BoardState
		You   struct {
			ID string
		}
	}
	jtest.RequireNil(t, json.NewDecoder(f).Decode(&req))

	youIDx := -1
	for i, snake := range req.Board.Snakes {
		if snake.ID == req.You.ID {
			youIDx = i
			break
		}
	}
	require.NotEqual(t, -1, youIDx)

	return &req.Board, youIDx
}
