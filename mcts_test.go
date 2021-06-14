package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func TestUCB1(t *testing.T) {
	// See https://www.youtube.com/watch?v=UXW2yZndl7U
	n := &mcnode{
		n: 2,
		t: 20,
		parent: &mcnode{n: 3},
	}
	val, _ := n.UCB1()
	require.Equal(t, 11.482303807367511, val)

	_, inf := (&mcnode{}).UCB1()
	require.True(t, inf)
}

func TestMCTSOnce(t *testing.T) {
	tests := []struct {
		Name string
		Exp map[string]float64
		Allow float64
		Ruleset rules.Ruleset
	}{
		//{
		//	Name: "testdata/input-015.json",
		//	Exp: map[string]float64{"down":-90, "left":-100, "right":-90, "up":-100},
		//	Allow: 15,
		//	Ruleset: &rules.SoloRuleset{},
		//},
		//{
		//	Name: "testdata/input-016.json",
		//	Exp: map[string]float64{"down":-100, "left":-100, "right":-100, "up":74},
		//	Allow: 5,
		//},
		//{
		//	Name: "testdata/input-017.json",
		//	Exp: map[string]float64{"down":-100, "left":-100, "right":-100, "up":-100},
		//},
		//{
		//	Name: "testdata/input-018.json",
		//	Exp: map[string]float64{"down":3, "left":-12, "right":-100, "up":-100},
		//	Allow: 5,
		//	Ruleset: &rules.SoloRuleset{},
		//},
		//{
		//	Name: "testdata/input-019.json",
		//	Exp: map[string]float64{"down":-100, "left":-100, "right":15, "up":0},
		//	Allow: 5,
		//	Ruleset: &rules.SoloRuleset{},
		//},
		{
			Name: "testdata/input-020.json",
			Exp: map[string]float64{"down":0, "left":-100, "right":-100, "up":-100},
			Allow: 5,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Ruleset == nil {
				test.Ruleset = &rules.StandardRuleset{}
			}

			{
				b, err := os.ReadFile(test.Name)
				jtest.RequireNil(t, err)

				var req GameRequest
				jtest.RequireNil(t, json.Unmarshal(b, &req))

				selectMCTS(context.Background(), req, basicWeights)
			}

			board, youIDx := fileToBoard(t, test.Name)
			root := &mcnode{
				Board:  board,
			}
			for _, move := range Moves {
				root.AddChild(board, nil, move.String())
			}

			scores := func(childs map[edge]*mcnode) map[string]float64{
				res := make(map[string]float64)
				for e, n := range childs {
					res[string(e)] = n.AvgScore()
				}
				return res
			}

			logd := func(msg string, args ...interface{}) {
				if true {
					return
				}
				if !strings.HasSuffix(msg, "\n") {
					msg += "\n"
				}
				fmt.Printf(msg, args...)
			}

			for i := 0; i < 10000; i++ {
				err := MCTSOnce(test.Ruleset, root, youIDx, logd)
				jtest.RequireNil(t, err)
				if root.dead {
					break
				}
			}
			logMCTS(time.Now(), root, root.childs["left"])
			score := scores(root.childs)
			for m, s := range test.Exp {
				require.True(t, s + test.Allow >= score[m], "%#v\nvs\n%#v", test.Exp, score)
				require.True(t, s - test.Allow <= score[m], "%#v\nvs\n%#v", test.Exp, score)
			}
		})
	}
}

func TestBack(t *testing.T) {
	board, _ := fileToBoard(t, "testdata/external/03.board.json")
	moves := []rules.SnakeMove{{ID: board.Snakes[0].ID, Move: Up.String()}}
	b  , err :=(&rules.StandardRuleset{}).CreateNextBoardState(board,moves )
	jtest.RequireNil(t, err)
	require.NotEmpty(t, b.Snakes[0].EliminatedCause)
}

func TestRollout(t *testing.T) {
	logd := func(string, ...interface{}){}

	board, youIDx := fileToBoard(t, "testdata/external/06.board.json")
	rand.Seed(3)
	score, err := RollOut(board, &rules.StandardRuleset{}, youIDx, "right", logd)
	jtest.RequireNil(t, err)
	require.Equal(t, 100.0, score)

	rand.Seed(100)
	score, err = RollOut(board, &rules.StandardRuleset{}, youIDx, "right", logd)
	jtest.RequireNil(t, err)
	require.Equal(t, 0.0, score)

	board, youIDx = fileToBoard(t, "testdata/external/07.board.json")
	rand.Seed(3)
	score, err = RollOut(board, &rules.StandardRuleset{}, youIDx, "up", logd)
	jtest.RequireNil(t, err)
	require.Equal(t, -100.0, score)

	board, youIDx = fileToBoard(t, "testdata/input-016.json")
	rand.Seed(3)
	score, err = RollOut(board, &rules.StandardRuleset{}, youIDx, "up", logd)
	jtest.RequireNil(t, err)
	require.Equal(t, 100.0, score)
}

func TestIsDeadly(t *testing.T) {
	tests := []struct {
		Name string
		SnakeIDx int
		Move Move
		Exp bool
	}{
		{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 1,
			Move: Left,
			Exp: true,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 1,
			Move: Right,
			Exp: true,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 1,
			Move: Up,
			Exp: false,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 1,
			Move: Down,
			Exp: false,
		},
		{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 0,
			Move: Left,
			Exp: true,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 0,
			Move: Right,
			Exp: false,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 0,
			Move: Up,
			Exp: true,
		},{
			Name: "testdata/external/06.board.json",
			SnakeIDx: 0,
			Move: Down,
			Exp: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			board, _ := fileToBoard(t, "testdata/external/06.board.json")
			require.Equal(t, test.Exp, isDeadlyRulesMove(board, test.SnakeIDx, test.Move.String()))
		})
	}
}

func fileToBoard(t *testing.T, file string) (*rules.BoardState, int) {
	f, err := os.Open(file)
	jtest.RequireNil(t, err)
	var req GameRequest
	jtest.RequireNil(t, json.NewDecoder(f).Decode(&req))

	return gameReqToBoard(req)

}
