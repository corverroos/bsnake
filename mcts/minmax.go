package mcts

import (
	"time"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/board"
	"github.com/corverroos/bsnake/heur"
)

type mx struct {
	move    string
	minimax float64
}

func Minimax(n *node, f *heur.Factors, hazards map[rules.Point]bool, ply int) ([]mx, error) {
	for _, moves := range board.GenMoveSet(n.board) {
		tup, err := genChild(n, moves)
		if err != nil {
			return nil, err
		}

		child := tup.child

		if child.board.Snakes[n.rootIdx].EliminatedCause == rules.EliminatedByOutOfHealth && child.depth > 20 {
			// Root snake out of health on deep node (uncertain). Rather not continue.
			continue
		}

		n.childs = append(n.childs, tup)

		if totals, ok, err := child.CheckTerminal(); err != nil {
			return nil, err
		} else if ok {
			child.termTotals = totals
			child.totals = totals
			child.n++
			continue
		}

		if ply == 1 {
			totals := heur.Calc(f, child.board, child.rootIdx, hazards)
			child.heurTotals = totals
			child.totals = totals
			child.n++
		} else {
			_, err := Minimax(child, f, hazards, ply-1)
			if err != nil {
				return nil, err
			}
		}
	}

	return MxPropagate(n), nil
}

func MxPropagate(n *node) []mx {

	res := make([]mx, len(n.board.Snakes))

	for i := 0; i < len(n.board.Snakes); i++ {
		var maxMove string
		var maxScore float64
		for _, move := range board.Moves {

			var min *float64
			for j := 0; j < len(n.childs); j++ {
				tup := &n.childs[j]
				if !tup.edge.Is(i, move) {
					continue
				}

				total := tup.child.totals[i]

				if min == nil || total < *min {
					min = &total
				}
			}

			if min != nil && (maxMove == "" || maxScore < *min) {
				maxMove = move
				maxScore = *min
			}
		}

		if maxMove != "" {
			res[i] = mx{
				move:    maxMove,
				minimax: maxScore,
			}
			n.heurTotals[i] = maxScore
			n.totals[i] = maxScore
		}
	}

	n.n++

	return res
}

func MxOnce(root *node, o *Opts, hazards map[rules.Point]bool) ([]mx, error) {
	n := selection(root, o)

	if !n.IsTerminal() {
		res, err := Minimax(n, o.HeurFactors, hazards, 1)
		if err != nil {
			return nil, err
		}
		if n.parent == nil {
			return res, nil
		}
	}

	for {
		n = n.parent
		res := MxPropagate(n)
		if n.parent == nil {
			return res, nil
		}
	}
}

func SelectMx(board *rules.BoardState, hazards []rules.Point, rootIDx int, o *Opts) (string, error) {
	t0 := time.Now()

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

	var moves []mx
	var err error
	for time.Since(t0) < time.Millisecond*340 {
		moves, err = MxOnce(root, o, hazmap)
		if err != nil {
			return "", err
		}
	}

	return moves[rootIDx].move, nil
}

func SelectMinimax(board *rules.BoardState, hazards []rules.Point, rootIDx int, f *heur.Factors, ply int) (string, error) {
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
