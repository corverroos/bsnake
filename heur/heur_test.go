package heur

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"
)

func TestLength(t *testing.T) {
	tests := []struct {
		Exp  []float64
		Lens []int
	}{
		{
			Exp: []float64{},
		},
		{
			Lens: []int{3, 3, 3},
			Exp:  []float64{0, 0, 0},
		},
		{
			Lens: []int{33, 13, 03},
			Exp:  []float64{0.34013605442176875, -0.06802721088435372, -0.27210884353741494},
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
		Name    string
		Control []float64
		Starve  []int
		Heur    []float64
		Move    string
	}{
		{
			Name:    "../testdata/input-001.json",
			Control: []float64{49},
			Starve:  []int{-1},
			Heur:    []float64{0.0002857142857142857},
			Move:    "left",
		},
		{
			Name:    "../testdata/input-006.json",
			Control: []float64{1, 120},
			Starve:  []int{0, -1},
			Heur:    []float64{-0.49714810442083174, 0.06939485766758494},
			Move:    "up",
		},
		{
			Name:    "../testdata/input-007.json",
			Control: []float64{91, 14, 16},
			Starve:  []int{-1, 0, -1},
			Heur:    []float64{0.08774471992653811, -0.018851239669421482, -0.06798438934802571},
			Move:    "up",
		}, {
			Name:    "../testdata/input-016.json",
			Control: []float64{8, 1},
			Starve:  []int{0, 0},
			Heur:    []float64{0.07011111111111111, -0.40244444444444444},
			Move:    "up",
		}, {
			Name:    "../testdata/input-017.json",
			Control: []float64{1, 8},
			Starve:  []int{0, 0},
			Heur:    []float64{-0.40244444444444444, 0.07011111111111111},
			Move:    "left",
		}, {
			Name:    "../testdata/input-022.json",
			Control: []float64{27, 40, 54},
			Starve:  []int{0, -1, -1},
			Heur:    []float64{-0.014836357303441934, 0.03235929831227638, -0.017068395554288966},
			Move:    "up",
		}, {
			Name:    "../testdata/input-029.json",
			Control: []float64{3, 22},
			Starve:  []int{0, 0},
			Heur:    []float64{-0.2649538461538462, 0.06555384615384616},
			Move:    "left",
		}, {
			Name:    "../testdata/input-030.json",
			Control: []float64{60, 61},
			Starve:  []int{-1, -1},
			Heur:    []float64{-0.035831320194956565, 0.03619495655859293},
			Move:    "up",
		}, {
			Name:    "../testdata/input-031.json",
			Control: []float64{81, 40},
			Starve:  []int{-1, -1},
			Heur:    []float64{-0.031156198347107443, 0.031610743801652894},
			Move:    "right",
		},
		{
			Name:    "../testdata/input-032.json",
			Control: []float64{95, 26},
			Starve:  []int{-1, -1},
			Heur:    []float64{0.03818801652892562, -0.037733471074380166},
			Move:    "up",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, youIdx := fileToBoard(t, test.Name)
			fmt.Printf("YouIdx: %d\n", youIdx)

			if strings.Contains(test.Name, "031") {
				b, _ = (&rules.StandardRuleset{}).CreateNextBoardState(b, []rules.SnakeMove{
					{ID: "gs_XhSkKctBXVSqxjkvKR6qkXXJ", Move: "left"},
					{ID: "gs_kyQbXRXC3879c4dRwBQt8kxV", Move: "right"},
				})
			}
			if strings.Contains(test.Name, "032") {
				var err error
				b, err = (&rules.StandardRuleset{}).CreateNextBoardState(b, []rules.SnakeMove{
					{ID: "gs_wgDwS8ckRBr4DmK7MFGjpW79", Move: "down"},
					{ID: "gs_FktVKX79vm8cYdRrxj6bWRv6", Move: "up"},
				})
				jtest.RequireNil(t, err)
			}

			control, starve := Flood(b, youIdx, nil)
			require.EqualValues(t, test.Control, control)
			require.EqualValues(t, test.Starve, starve)

			f := &Factors{
				Control: 0.05,
				Length:  0.4,
				Boxed:   -0.5,
				Hunger:  -0.001,
				Starve:  -0.9,
				Walls:   0.001,
			}

			heur := Calc(f, b, youIdx, nil)
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
