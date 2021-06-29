package mcts

import (
	"fmt"
	"testing"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"

	"github.com/corverroos/bsnake/heur"
)

func TestMinimax2(t *testing.T) {
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
		//{
		//	Name: "../testdata/input-021.json",
		//	Exp:  "left",
		//}, // This is random up or left
		{
			Name: "../testdata/input-022.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-023.json",
			Exp:  "right",
		},
		{
			Name: "../testdata/input-024.json",
			Exp:  "left",
		},
		{
			Name: "../testdata/input-025.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-027.json",
			Exp:  "down",
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

			opts := &OptsV3
			opts.logd = func(s string, i ...interface{}) {
				//fmt.Printf(s+"\n", i...)
			}

			{
				root := NewRoot(rulset, board, rootIdx)
				fmt.Printf("rootIdx=%v\n", rootIdx)

				f := &heur.Factors{
					Control: 0.5,
					Length:  0.3,
					Starve:  -0.9,
				}

				mxl, err := Minimax(root, f, nil, 2)
				jtest.RequireNil(t, err)

				for idx, mx := range mxl {
					fmt.Printf("snake=%v move=%v minimax=%v\n", idx, mx.move, mx.minimax)
				}
				require.Equal(t, test.Exp, mxl[rootIdx].move)
			}

			{
				root := NewRoot(rulset, board, rootIdx)
				fmt.Printf("rootIdx=%v\n", rootIdx)

				f := &heur.Factors{
					Control: 0.5,
					Length:  0.3,
					Starve:  -0.9,
				}

				var mxl map[int]mx
				for i := 0; i < 20; i++ {
					var err error
					mxl, err = MinimaxOnce(root, f, opts, nil)
					jtest.RequireNil(t, err)
				}

				for idx, mx := range mxl {
					fmt.Printf("snake=%v move=%v minimax=%v\n", idx, mx.move, mx.minimax)
				}
				require.Equal(t, test.Exp, mxl[rootIdx].move)
			}

		})
	}
}
