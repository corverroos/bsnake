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
			Exp:  "up",
		},
		{
			Name: "../testdata/input-027.json",
			Exp:  "right",
		},
		{
			Name: "../testdata/input-028.json",
			Exp:  "up",
		},
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
			opts := &OptsV4
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
			for i := 0; i < 10000; i++ {
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

			//require.Equal(t, test.Exp, root.RobustSafeMove(rootIdx))

			opts.logr(root, rootIdx, "")

			if !strings.Contains(t.Name(), "-021") && !strings.Contains(t.Name(), "-027") {
				//require.Equal(t, test.Exp, root.MinMaxMove(rootIdx))
				//require.Equal(t, test.Exp, root.RobustMoves(rootIdx)[0])
			}

		})
	}

	//fmt.Printf("JCR: totals=%v\n", totals)
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
	res := make(map[int]float64)
	assignLenRewards(&Opts{}, res, map[int]int{0: 0, 1: 0}, map[int]int{0: 2, 1: 0})
	require.Equal(t, 0.1, res[0])
	require.Equal(t, -0.1, res[1])

	res = make(map[int]float64)
	assignLenRewards(&Opts{Version: 3}, res, map[int]int{0: 0, 1: 0}, map[int]int{0: 2, 1: 0})
	require.Equal(t, 0.8, res[0])
	require.Equal(t, -0.3, res[1])
}

//func TestAssignLenTups(t *testing.T) {
//	tests := []struct {
//		Count int
//		Exp map[int]float64
//	}{
//		{
//			Count: 0,
//			Exp: map[int]float64{},
//		},{
//			Count: 1,
//			Exp: map[int]float64{0:0.5},
//		},{
//			Count: 2,
//			Exp: map[int]float64{0:0.5, 1:-0.5},
//		},{
//			Count: 3,
//			Exp: map[int]float64{0:0.5, 1:0, 2:-0.5},
//		},{
//			Count: 6,
//			Exp: map[int]float64{0:0.5, 1:0.3333333333333333, 2:0.16666666666666666, 3:-0.16666666666666666, 4:-0.3333333333333333, 5:-0.5},
//		},
//	}
//
//	for _, test := range tests {
//		t.Run(fmt.Sprint(test.Count), func(t *testing.T) {
//			res := make(map[int]float64)
//			var tups []intTup
//			for i := 0; i < test.Count; i++ {
//				tups = append(tups, intTup{
//					K: i,
//					V: 10-i,
//				})
//			}
//			assignLenRewards(res, tups)
//			require.EqualValues(t, res, test.Exp)
//		})
//	}
//}

func TestGenMoves(t *testing.T) {
	tests := []struct {
		Name string
		Exp  []map[int]string
	}{
		{
			Name: "../testdata/input-021.json",
			Exp: []map[int]string{
				{0: "down", 1: "up"},
				{0: "right", 1: "up"},
				{0: "down", 1: "left"},
				{0: "right", 1: "left"},
			},
		}, {
			Name: "../testdata/input-022.json",
			Exp: []map[int]string{
				{0: "right", 1: "down", 2: "up"},
				{0: "right", 1: "right", 2: "up"},
				{0: "right", 1: "down", 2: "right"},
				{0: "right", 1: "right", 2: "right"},
				{0: "right", 1: "down", 2: "left"},
				{0: "right", 1: "right", 2: "left"}},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, _ := fileToBoard(t, test.Name)

			totals := board.GenMoveSet(b)
			require.EqualValues(t, test.Exp, totals)
		})
	}
}

func TestPlayoutRational(t *testing.T) {
	tests := []struct {
		Name string
		Exp  map[int]float64
	}{
		{
			Name: "../testdata/input-021.json",
			Exp:  map[int]float64{0: 1, 1: -1},
		}, {
			Name: "../testdata/input-022.json",
			Exp:  map[int]float64{0: -1, 1: 1, 2: -1},
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
		n.UpdateScores(map[int]float64{0: f})
	}
	require.Equal(t, 30.0, n.ScoreVariance(0))

	n = &node{}
	for _, f := range []float64{1, 2, 2, 4, 6} {
		n.UpdateScores(map[int]float64{0: f})
	}
	require.Equal(t, 4.0, n.ScoreVariance(0))
}

func TestChildMoves(t *testing.T) {
	board, rootIdx := fileToBoard(t, "../testdata/input-021.json")
	require.Equal(t, 1, rootIdx)

	n0 := NewRoot(&rules.StandardRuleset{}, board, rootIdx)
	require.Zero(t, n0.AvgScore(rootIdx))

	moves := map[int]string{0: "left", 1: "right", 2: "down"}
	n1, err := n0.AppendChild(moves)
	jtest.RequireNil(t, err)

	require.Len(t, n0.childs, 1)
	require.NotEqual(t, n0.board, n1.board)

	for _, tuple := range n0.childs {
		e := tuple.edge
		c := tuple.child
		require.Equal(t, "0b10100011", fmt.Sprintf("%#b", e))
		require.Equal(t, "lrd", e.String())
		require.True(t, e.Is(0, moves[0]))
		require.True(t, e.Is(1, moves[1]))
		require.True(t, e.Is(2, moves[2]))
		require.False(t, e.Is(3, moves[3]))

		ucdb1, inf := c.UCB1(rootIdx)
		require.Zero(t, ucdb1)
		require.True(t, inf)
	}
}

func TestEdge(t *testing.T) {
	e := newEdge(map[int]string{0: "right"})
	require.Equal(t, "0b100", fmt.Sprintf("%#b", e))
	require.Equal(t, "r", e.String())
	require.True(t, e.Is(0, "right"))
	require.False(t, e.Is(0, "left"))
	require.False(t, e.Is(1, "right"))

	e = newEdge(map[int]string{0: "up", 2: "up", 3: "left"})
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
