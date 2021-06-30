package mcts

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/board"
	"github.com/corverroos/bsnake/heur"
)

//var totals = map[string]time.Duration{}
//
//var lat = func(name string) func() {
//	t0 := time.Now()
//	return func() {
//		totals[name] += time.Since(t0)
//	}
//}

func Once(root *node, o *Opts) error {
	node := selection(root, o)
	o.Logd("selected depth=%d", node.depth)

	if node.IsTerminal() {
		o.Logd("propagate old terminal")
		propagation(node, node.termTotals)
		return nil
	}

	if node.n == 1 {
		var err error
		node, err = expansion(node, o)
		if err != nil {
			return err
		}
	}

	if node.n != 0 {
		panic("playout visited node")
	}

	if totals, ok, err := node.CheckTerminal(); err != nil {
		return err
	} else if ok {
		o.Logd("propagate new terminal, totals=%v", totals)
		node.termTotals = totals
		propagation(node, node.termTotals)
		return nil
	}

	var totals []float64
	var err error
	if o.LeafPlayout {
		totals, err = playoutRandomRational(root, node, o)
		if err != nil {
			return err
		}
		o.Logd("propagate play-out, totals=%v", totals)
	} else if o.LeafHeur {
		totals = heur.Calc(o.HeurFactors, node.board, o.hazards)
		o.Logd("propagate heuristics, totals=%v", totals)
	} else {
		panic("invalid options, no leaf strategy")
	}

	propagation(node, totals)

	return nil
}

// expansion adds all rational move child nodes to n and returns the first.
func expansion(n *node, o *Opts) (*node, error) {
	//defer lat("expansion")()
	if n.IsTerminal() {
		return n, nil
	}

	var res *node
	moveSet := board.GenMoveSet(n.board)

	if o.AvoidLH2H {
		temp := make([][]string, 0, len(moveSet))
		for _, moves := range moveSet {
			if board.IsLoosingH2H(n.board, n.rootIdx, moves[n.rootIdx]) {
				continue
			}
			temp = append(temp, moves)
		}
		if len(temp) > 0 {
			moveSet = temp
		}
	}

	for i, moves := range moveSet {
		child, err := n.AppendChild(moves)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			res = child
		}
	}

	o.Logd("expanded sets=%d select edge=%s", len(moveSet), newEdge(moveSet[0]))

	if res == nil {
		panic("no child node and not terminal")
	}

	return res, nil
}

func playoutRandomRational(root, node *node, o *Opts) ([]float64, error) {
	//defer lat("playout")()
	l := len(root.idsByIdx)
	b := node.board
	r := node.ruleset

	maxcount := o.MaxPlayout
	if len(root.board.Snakes) == 1 {
		maxcount = 100
	}

	randMoves := func(b *rules.BoardState) []rules.SnakeMove {
		res := make([]rules.SnakeMove, l)
		for i := 0; i < l; i++ {
			if b.Snakes[i].EliminatedCause != "" {
				continue
			}
			for j, move := range board.RandMoves() {
				if j < 3 && !board.IsRationalMove(b, i, move) {
					continue
				}
				res[i] = rules.SnakeMove{
					ID:   b.Snakes[i].ID,
					Move: move,
				}
				break
			}
		}
		return res
	}

	greedyMoves := func(b *rules.BoardState) []rules.SnakeMove {
		var res []rules.SnakeMove
		for i := 0; i < l; i++ {
			res = append(res, rules.SnakeMove{
				ID:   b.Snakes[i].ID,
				Move: o.GreedyHeur(b, i),
			})
		}
		return res
	}

	startLens := make([]int, l)
	for i := 0; i < len(root.board.Snakes); i++ {
		startLens[i] = len(root.board.Snakes[i].Body)
	}

	var count int
	res := make([]float64, l)
	for {
		var err error

		moveFunc := randMoves
		if rand.Float64() < o.GreedyProb {
			moveFunc = greedyMoves
		}

		moves := moveFunc(b)

		b, err = r.CreateNextBoardState(b, moves)
		if err != nil {
			return nil, err
		}

		count++

		over, err := r.IsGameOver(b)
		if err != nil {
			return nil, err
		}

		if count < maxcount && !over {
			continue
		}

		for i := 0; i < l; i++ {
			if b.Snakes[i].EliminatedCause != "" {
				res[i] = -1
				continue
			}
			if over {
				res[i] = 1
				continue
			}
		}

		if over {
			return res, nil
		}

		if o.PlayoutMaxHeur {
			return heur.Calc(o.HeurFactors, b, o.hazards), nil
		}

		endLens := make([]int, l)
		for i := 0; i < l; i++ {
			endLens[i] = len(b.Snakes[i].Body)
		}
		assignLenRewards(o, res, startLens, endLens)
		return res, nil
	}
}

func assignAreaReqards(res map[int]float64, b *rules.BoardState) {
	if len(b.Snakes) == 1 {
		return
	}

	sums := SumVoronoi(b)

	var total float64
	for _, v := range sums {
		total += float64(v)
	}

	avail := 0.5
	for i, v := range sums {
		res[i] = (avail * float64(v) / total) - avail/2
	}
}

func SumVoronoi(b *rules.BoardState) map[int]int {
	res := make(map[int]int)
	ycmin := b.Height / 3
	xcmin := b.Width / 3
	ycmax := b.Height - ycmin
	xcmax := b.Width - xcmin

	for x := int32(0); x < b.Width; x++ {
		for y := int32(0); y < b.Height; y++ {
			minDist := b.Width * b.Height
			var minS int
			for s := 0; s < len(b.Snakes); s++ {
				for i, c := range b.Snakes[s].Body {
					if i > len(b.Snakes[s].Body)*3*2 {
						// Only consider first 2/3 of snake
						break
					}

					distX := c.X - x
					distY := c.Y - y
					if distX < 0 {
						distX = -distX
					}
					if distY < 0 {
						distY = -distY
					}
					dist := distX + distY
					if dist < minDist {
						minDist = dist
						minS = s
					}
					if dist == 0 {
						break
					}
				}
			}
			res[minS]++

			if y >= ycmin && y < ycmax && x >= xcmin && x < xcmax {
				// Double points for controlling the centre
				res[minS]++
			}
		}
	}
	return res
}

func assignLenRewards(o *Opts, res []float64, start, end []int) {
	l := len(start)

	if l == 1 {
		res[0] = -0.1 * math.Min(float64(end[0]-start[0]), 8)
		return
	}

	rank := func(m []int, i int) float64 {
		var rank float64
		myLen := m[i]
		for j, l := range m {
			if j == i {
				continue
			}
			if l >= myLen {
				rank++
			}
		}
		return rank
	}

	if o.Version == 3 {
		for i := 0; i < l; i++ {
			if res[i] != 0 {
				continue
			}

			// Snakes get 0.1 for each food they eat
			delta := -0.3 + 0.3*float64(end[i]-start[i])
			if rank(end, i) == 0 {
				// Longest snake at the end gets 0.5
				delta += 0.5
			}

			res[i] = math.Min(delta, 0.9)
		}
	} else {
		for i := 0; i < l; i++ {
			if rank(end, i) == 0 {
				// Longest snake at the end gets 0.5
				res[i] = 0.1 * float64(len(end)-1)
				continue
			}
			res[i] -= 0.1
			// Other snakes get 0.1 for each food they eat
			res[i] += 0.2 * float64(end[i]-start[i])
		}
	}
}

func selection(root *node, o *Opts) *node {
	n := root
	for {
		if n.IsLeaf() {
			return n
		}

		if n.n < o.SelectRandom {
			random := n.childs[rand.Intn(len(n.childs))]
			n = random.child
			o.Logd("select random child, depth=%d, edge=%s", n.depth, random.edge)
			continue
		}

		type stats struct {
			sumN       float64
			sumTotals  float64
			sumSquares float64
			heuristic  float64
		}

		allStats := make([][]stats, len(n.idsByIdx))

		// Calculate stats for each move for each snake
		for _, tuple := range n.childs {
			if tuple.child.n == 0 {
				o.Logd("select unexplored child, depth=%d, edge=%s", tuple.child.depth, tuple.edge)
				return tuple.child
			}

			if o.SelectHeur && len(tuple.child.heurTotals) == 0 {
				tuple.child.heurTotals = heur.Calc(o.HeurFactors, tuple.child.board, o.hazards)
			}

			for i := 0; i < len(n.idsByIdx); i++ {
				snakeStats := allStats[i]
				if snakeStats == nil {
					snakeStats = make([]stats, 4)
					allStats[i] = snakeStats
				}

				for midx, move := range board.Moves {
					if !tuple.edge.Is(i, move) {
						continue
					}
					st := snakeStats[midx]
					st.sumN += tuple.child.n
					st.sumTotals += tuple.child.totals[i]
					st.sumSquares += tuple.child.totalSquares[i]
					st.heuristic += tuple.child.heurTotals[i]
					snakeStats[midx] = st
					break
				}
			}
		}

		maxMoves := make([]string, len(n.idsByIdx))
		for i := 0; i < len(n.idsByIdx); i++ {
			var max *float64
			for midx, move := range board.Moves {
				st := allStats[i][midx]
				if st.sumN == 0 {
					continue
				}

				c := o.UCB1_C
				if o.Tuned && st.sumN > 1 {
					// UCB1-Tuned: https://dke.maastrichtuniversity.nl/m.winands/documents/sm-tron-bnaic2013.pdf
					variance := (st.sumSquares - (st.sumTotals*st.sumTotals)/st.sumN) / (st.sumN - 1)

					v := variance + math.Sqrt(2*math.Log(n.n)/st.sumN)
					c *= math.Min(0.25, v)
				}

				ucb1 := st.sumTotals/st.sumN + math.Sqrt(c*math.Log(n.n)/st.sumN) + st.heuristic/(st.sumN+1)
				if max == nil || ucb1 > *max {
					max = &ucb1
					maxMoves[i] = move
				}
			}
		}

		maxEdge := newEdge(maxMoves)
		var next *node
		for _, tuple := range n.childs {
			if tuple.edge == maxEdge {
				next = tuple.child
				break
			}
		}

		if next == nil {
			var edges []string
			for _, tuple := range n.childs {
				edges = append(edges, tuple.edge.String())
			}
			panic(fmt.Sprintf("missing child: %s not in %v, ", maxEdge, edges))
		}
		n = next
		o.Logd("select DUCT child, depth=%d, edge=%s", n.depth, maxEdge)
	}
}

func propagation(node *node, totals []float64) {
	n := node
	for n != nil {
		for idx, t := range totals {
			n.totals[idx] += t
			n.totalSquares[idx] += t * t
		}
		n.n++
		n = n.parent
	}
}

func SelectMove(ctx context.Context, board *rules.BoardState, hazards []rules.Point, rootIDx int, o *Opts) (string, error) {
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

	o.hazards = make(map[rules.Point]bool)
	for _, hazard := range hazards {
		o.hazards[hazard] = true
	}

	root := NewRoot(ruleset, board, rootIDx)

	for time.Since(t0) < time.Millisecond*340 {
		err := Once(root, o)
		if err != nil {
			return "", err
		}
	}

	var move string
	if o.Version == 1 {
		move = root.RobustMoves(rootIDx)[0]
	} else {
		move = root.RobustSafeMove(rootIDx)
	}

	o.LogResults(root, rootIDx, move)

	return move, nil
}
