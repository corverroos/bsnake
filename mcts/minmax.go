package mcts

import (
	"fmt"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/board"
	"github.com/corverroos/bsnake/heur"
)

type mx struct {
	move    string
	minimax float64
}

var debug = 0

func Minimax(n *node, f heur.Factors, hazards map[rules.Point]bool, ply int) (map[int]mx, error) {
	em := make(map[edge]map[int]float64)
	for _, moves := range board.GenMoveSet(n.board) {
		tup, err := genChild(n, moves)
		if err != nil {
			return nil, err
		}

		child := tup.child

		if totals, ok, err := child.CheckTerminal(); err != nil {
			return nil, err
		} else if ok {
			em[tup.edge] = totals
			continue
		}

		if ply == 1 {
			em[tup.edge] = heur.Calc(f, child.board, hazards)
		} else {
			mxl, err := Minimax(child, f, hazards, ply-1)
			if err != nil {
				return nil, err
			}
			totals := make(map[int]float64)
			for idx, mx := range mxl {
				totals[idx] = mx.minimax
			}
			em[tup.edge] = totals
		}
	}

	res := make(map[int]mx)

	for i := 0; i < len(n.board.Snakes); i++ {
		var maxMove string
		var maxScore float64
		for _, move := range board.Moves {

			var min *float64
			for e, m := range em {
				if !e.Is(i, move) {
					continue
				}

				total := m[i]
				if ply == debug {
					fmt.Printf("ply=%d snake=%v move=%v edge=%s minimax=%v\n", ply, i, move, e, total)
				}
				if min == nil || total < *min {
					min = &total
				}
			}
			if ply == debug && min != nil {
				fmt.Printf("move=%s min=%v\n", move, *min)
			} else if ply == debug {
				fmt.Printf("move=%s no min\n", move)
			}
			if min != nil && (maxMove == "" || maxScore < *min) {
				maxMove = move
				maxScore = *min
			}
		}
		if ply == debug {
			fmt.Printf("maxMove=%v, maxScore=%v \n", maxMove, maxScore)
		}
		if maxMove != "" {
			res[i] = mx{
				move:    maxMove,
				minimax: maxScore,
			}
		}
	}

	return res, nil
}

func SelectMinimax(board *rules.BoardState, hazards []rules.Point, rootIDx int, f heur.Factors, ply int) (string, error) {
	var ruleset rules.Ruleset
	if len(board.Snakes) == 1 {
		ruleset = &rules.SoloRuleset{}
	} else if len(hazards) > 0 {
		ruleset = &RoyaleRuleset{
			Hazards: hazards,
		}
	} else {
		ruleset = &rules.StandardRuleset{}
	}

	hazmap := make(map[rules.Point]bool)
	for _, hazard := range hazards {
		hazmap[hazard] = true
	}

	root := NewRoot(ruleset, board, rootIDx)

	res, err := Minimax(root, f, hazmap, ply)
	if err != nil {
		return "", err
	}

	return res[rootIDx].move, nil
}
