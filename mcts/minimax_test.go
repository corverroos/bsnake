package mcts

import (
	"fmt"
	"testing"

	"github.com/BattlesnakeOfficial/rules"
	"github.com/luno/jettison/jtest"
	"github.com/stretchr/testify/require"

	"github.com/corverroos/bsnake/board"
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
			Exp:  "left",
		},
		{
			Name: "../testdata/input-028.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-033.json",
			Exp:  "up",
		},
		{
			Name: "../testdata/input-034.json",
			Exp:  "down",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			b, rootIdx := fileToBoard(t, test.Name)

			rulset := rules.Ruleset(&rules.StandardRuleset{})
			if len(b.Snakes) == 1 {
				rulset = &rules.SoloRuleset{}
			}

			opts := &OptsV2
			opts.logd = func(s string, i ...interface{}) {
				//fmt.Printf(s+"\n", i...)
			}

			root := NewRoot(rulset, b, rootIdx)
			fmt.Printf("rootIdx=%v\n", rootIdx)

			var mxl []mx
			for i := 0; i < 10000; i++ {
				var err error
				mxl, err = MxOnce(root, opts, nil)
				jtest.RequireNil(t, err)
			}

			for idx, mx := range mxl {
				fmt.Printf("snake=%v move=%v minimax=%v\n", idx, mx.move, mx.minimax)
			}

			n := root
			//for _, e := range []edge{
			//	 m2e(d,d),
			//	//m2e(u,r,l),
			//	//m2e(u,u,l),
			//	//m2e(u,u,d),
			//	//m2e(l,r,r),
			//} {
			//	var found bool
			//	for _, tup := range n.childs {
			//		if tup.edge == e {
			//			n = tup.child
			//			found = true
			//		}
			//	}
			//	if !found {
			//		panic("not found")
			//	}
			//	fmt.Printf("edge=%v\n%s", e, board.PrintBoard(n.board, nil))
			//}

			//n := root
			for i := 0; i < len(root.idsByIdx); i++ {
				for _, move := range board.Moves {
					var min *float64
					for _, tup := range n.childs {
						if !tup.edge.Is(i, move) {
							continue
						}
						tot := tup.child.totals[i]
						if min == nil || tot < *min {
							min = &tot
						}
						fmt.Printf("e  snake=%v move=%v edge=%v total=%v\n", i, move, tup.edge.String(), tot)
					}
					if min == nil {

					} else {
						fmt.Printf("m   snake=%v move=%v min=%v\n", i, move, *min)
					}
				}
			}

			require.Equal(t, test.Exp, mxl[rootIdx].move)
		})
	}
}

const u, d, l, r = 0, 1, 2, 3

func m2e(ml ...int) edge {
	var res []string
	for _, m := range ml {
		switch m {
		case 0:
			res = append(res, "up")
		case 1:
			res = append(res, "down")
		case 2:
			res = append(res, "left")
		case 3:
			res = append(res, "right")
		}
	}
	return newEdge(res)
}
