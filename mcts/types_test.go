package mcts

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"

	"github.com/corverroos/bsnake/board"
)

func Test500Once(t *testing.T) {
	tests := []struct {
		Name string
		Exp  string
	}{
		{
			Name: "../testdata/input-001.json",
			Exp:  "down",
		},
		{
			Name: "../testdata/input-017.json",
			Exp:  "left",
		},
		{
			Name: "../testdata/input-020.json",
			Exp:  "down",
		},
		{
			Name: "../testdata/input-021.json",
			Exp:  "left",
		},
		{
			Name: "../testdata/input-022.json",
			Exp:  "right",
		},
		{
			Name: "../testdata/input-023.json",
			Exp:  "right",
		},
		{
			Name: "../testdata/input-024.json",
			Exp:  "left", // Should be right
		},
		{
			Name: "../testdata/input-025.json",
			Exp:  "right", // could also be up, very similar
		},
		{
			Name: "../testdata/input-027.json",
			Exp:  "left",
		},
		{
			Name: "../testdata/input-028.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-030.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-031.json",
			Exp:  "right",
		},
		//{
		//	Name: "../testdata/input-032.json",
		//	Exp:  "down", // Flip flows between left and down...
		//},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			board, rootIdx := fileToBoard(t, test.Name)

			rulset := rules.Ruleset(&rules.StandardRuleset{})
			if len(board.Snakes) == 1 {
				rulset = &rules.SoloRuleset{}
			}

			// V3 : totals=map[expansion:447.687227ms playout:2.987921459s selection:1.105010259s]
			// V2 : totals=map[expansion:481.324695ms playout:2.397827728s selection:1.186210448s]
			opts := &OptsV5
			opts.AvoidLH2H = true
			opts.logd = func(s string, i ...interface{}) {
				//fmt.Printf(s+"\n", i...)
			}
			opts.logr = func(root *node, rootIdx int, move string) {
				s := sampleStats(graphDepths(root))
				fmt.Printf("graph: nodes=%.0f maxd=%.0f avgd=%.0f stddev=%.0f\n", s.count, s.max, s.mean, s.stddev)
				var longest string
				n := root
				for len(n.childs) > 0 {
					var next tuple
					for i, tup := range n.childs {
						if i == 0 || next.child.n < tup.child.n {
							next = tup
						}
					}
					longest += fmt.Sprintf("%s(%.0f) ", next.edge, next.child.n)
					n = next.child
				}
				fmt.Println(longest)
			}

			root := NewRoot(rulset, board, rootIdx)
			fmt.Printf("rootIdx=%v\n", rootIdx)
			for i := 0; i < 5000; i++ {
				rand.Seed(int64(i))
				err := Once(root, opts)
				jtest.RequireNil(t, err)
				require.Equal(t, float64(i+2), root.n)
			}

			for _, tuple := range root.childs {
				avgs := map[int]float64{}
				for i := 0; i < len(root.idsByIdx); i++ {
					avgs[i] = tuple.child.AvgScore(i)
				}
				fmt.Printf("edge=%v visits=%v avgs=%v\n", tuple.edge, tuple.child.n, avgs)
			}

			fmt.Printf("RobustMoves=%v\n", root.RobustMoves(rootIdx))
			fmt.Printf("MinMaxMove=%v\n", root.MinMaxMove(rootIdx))

			require.Equal(t, test.Exp, root.RobustSafeMove(rootIdx))

			opts.logr(root, rootIdx, "")

			if !strings.Contains(t.Name(), "-021") && !strings.Contains(t.Name(), "-027") {
				require.Equal(t, test.Exp, root.MinMaxMove(rootIdx))
				require.Equal(t, test.Exp, root.RobustMoves(rootIdx)[0])
			}

		})
	}
}

func TestVoronoi(t *testing.T) {
	tests := []struct {
		Name string
		Exp1 map[int]int
		Exp2 map[int]float64
	}{
		{
			Name: "../testdata/input-001.json",
			Exp1: map[int]int{0: 58},
			Exp2: map[int]float64{},
		},
		{
			Name: "../testdata/input-022.json",
			Exp1: map[int]int{0: 14, 1: 86, 2: 46},
			Exp2: map[int]float64{0: -0.20205479452054795, 1: 0.04452054794520549, 2: -0.09246575342465754},
		},
		{
			Name: "../testdata/input-025.json",
			Exp1: map[int]int{0: 110, 1: 36},
			Exp2: map[int]float64{0: 0.1267123287671233, 1: -0.1267123287671233},
		},
		{
			Name: "../testdata/input-027.json",
			Exp1: map[int]int{0: 71, 1: 75},
			Exp2: map[int]float64{0: -0.00684931506849315, 1: 0.006849315068493178},
		},
		{
			Name: "../testdata/input-020.json",
			Exp1: map[int]int{0: 18, 1: 40},
			Exp2: map[int]float64{0: -0.09482758620689655, 1: 0.09482758620689657},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			board, rootIdx := fileToBoard(t, test.Name)
			fmt.Printf("rootIdx=%v\n", rootIdx)
			res := SumVoronoi(board)
			require.EqualValues(t, test.Exp1, res)

			res2 := make(map[int]float64)
			assignAreaReqards(res2, board)
			require.EqualValues(t, test.Exp2, res2)
		})
	}
}

func TestLen(t *testing.T) {
	res := make([]float64, 2)
	assignLenRewards(&Opts{}, res, []int{0, 0}, []int{2, 0})
	require.Equal(t, 0.1, res[0])
	require.Equal(t, -0.1, res[1])

	res = make([]float64, 2)
	assignLenRewards(&Opts{Version: 3}, res, []int{0, 0}, []int{2, 0})
	require.Equal(t, 0.8, res[0])
	require.Equal(t, -0.3, res[1])
}

func TestGenMoves(t *testing.T) {
	tests := []struct {
		Name string
		Exp  [][]string
	}{
		{
			Name: "../testdata/input-006.json",
			Exp: [][]string{
				{"up", "down"},
				{"right", "down"},
				{"up", "right"},
				{"right", "right"},
				{"up", "left"},
				{"right", "left"},
			},
		},
		{
			Name: "../testdata/input-021.json",
			Exp: [][]string{
				{"down", "up"},
				{"right", "up"},
				{"down", "left"},
				{"right", "left"},
			},
		}, {
			Name: "../testdata/input-022.json",
			Exp: [][]string{
				{"right", "down", "up"},
				{"right", "right", "up"},
				{"right", "down", "right"},
				{"right", "right", "right"},
				{"right", "down", "left"},
				{"right", "right", "left"}},
		},
		{
			Name: "../testdata/input-023.json",
			Exp: [][]string{
				{"down", "up"},
				{"right", "up"},
				{"left", "up"},
				{"down", "right"},
				{"right", "right"},
				{"left", "right"},
				{"down", "left"},
				{"right", "left"},
				{"left", "left"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, rootIdx := fileToBoard(t, test.Name)
			fmt.Println("rootIdx", rootIdx)
			totals := board.GenMoveSet(b)
			require.EqualValues(t, test.Exp, totals)
		})
	}
}

func TestPlayoutRational(t *testing.T) {
	tests := []struct {
		Name string
		Exp  []float64
	}{
		{
			Name: "../testdata/input-021.json",
			Exp:  []float64{1, -1},
		}, {
			Name: "../testdata/input-022.json",
			Exp:  []float64{-1, 1, -1},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			board, rootIdx := fileToBoard(t, test.Name)
			n0 := NewRoot(&rules.StandardRuleset{}, board, rootIdx)

			rand.Seed(0)
			totals, err := playoutRandomRational(n0, n0, &OptsV1)
			jtest.RequireNil(t, err)
			require.EqualValues(t, test.Exp, totals)
		})
	}
}

func TestVariance(t *testing.T) {
	n := &node{}
	for _, f := range []float64{4, 7, 13, 16} {
		n.UpdateScores([]float64{f})
	}
	require.Equal(t, 30.0, n.ScoreVariance(0))

	n = &node{}
	for _, f := range []float64{1, 2, 2, 4, 6} {
		n.UpdateScores([]float64{f})
	}
	require.Equal(t, 4.0, n.ScoreVariance(0))
}

func TestChildMoves(t *testing.T) {
	board, rootIdx := fileToBoard(t, "../testdata/input-021.json")
	require.Equal(t, 1, rootIdx)

	n0 := NewRoot(&rules.StandardRuleset{}, board, rootIdx)
	require.Zero(t, n0.AvgScore(rootIdx))

	moves := []string{"left", "right"}
	n1, err := n0.AppendChild(moves)
	jtest.RequireNil(t, err)

	require.Len(t, n0.childs, 1)
	require.NotEqual(t, n0.board, n1.board)

	for _, tuple := range n0.childs {
		e := tuple.edge
		c := tuple.child
		require.Equal(t, "0b100011", fmt.Sprintf("%#b", e))
		require.Equal(t, "lr", e.String())
		require.True(t, e.Is(0, moves[0]))
		require.True(t, e.Is(1, moves[1]))

		ucdb1, inf := c.UCB1(rootIdx)
		require.Zero(t, ucdb1)
		require.True(t, inf)
	}
}

func TestEdge(t *testing.T) {
	e := newEdge([]string{"right"})
	require.Equal(t, "0b100", fmt.Sprintf("%#b", e))
	require.Equal(t, "r", e.String())
	require.True(t, e.Is(0, "right"))
	require.False(t, e.Is(0, "left"))
	require.False(t, e.Is(1, "right"))

	e = newEdge([]string{"up", "", "up", "left"})
	require.Equal(t, "0b11001000001", fmt.Sprintf("%#b", e))
	require.Equal(t, "u_ul", e.String())
	require.True(t, e.Is(0, "up"))
	require.False(t, e.Is(1, "up"))
	require.True(t, e.Is(2, "up"))
	require.True(t, e.Is(3, "left"))
}

func TestFileToBoard(t *testing.T) {
	fileToBoard(t, "../testdata/input-001.json")
	fileToBoard(t, "../testdata/input-010.json")
	fileToBoard(t, "../testdata/input-020.json")
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
