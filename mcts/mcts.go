package mcts

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/BattlesnakeOfficial/rules"
)

type logd = func(string, ...interface{})

func Once(root *node, logd logd) error {
	node := selection(root, logd)
	logd("selected depth=%d", node.depth)

	if node.IsTerminal() {
		logd("propagate old terminal")
		propagation(node, node.termTotals)
		return nil
	}

	if node.n == 1 {
		var err error
		node, err = expansion(node, logd)
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
		logd("propagate new terminal, totals=%v", totals)
		node.termTotals = totals
		propagation(node, node.termTotals)
		return nil
	}

	t0 := time.Now()
	totals, err := playoutRandomRational(node)
	if err != nil {
		return err
	}
	logd("propagate play-out, totals=%v, latency=%dus", totals, time.Since(t0).Microseconds())
	propagation(node, totals)

	return nil
}

// expansion adds all rational move child nodes to n and returns the first.
func expansion(n *node, logd logd) (*node, error) {
	if n.IsTerminal() {
		return n, nil
	}

	var res *node
	moveSet := genMoveSet(n.board)
	for i, moves := range moveSet {
		child, err := n.GenChild(moves)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			res = child
		}
	}

	logd("expanded sets=%d select edge=%s", len(moveSet), newEdge(moveSet[0]))

	if res == nil {
		panic("no child node and not terminal")
	}

	return res, nil
}

func playoutRandomRational(node *node) (map[int]float64, error) {
	b := node.board
	r := node.ruleset

	randMoves := func(b *rules.BoardState) []rules.SnakeMove {
		res := make([]rules.SnakeMove, len(b.Snakes))
		for i := 0; i < len(b.Snakes); i++ {
			if b.Snakes[i].EliminatedCause != "" {
				continue
			}
			for j, move := range RandMoves() {
				if j < 3 && !isRationalMove(b, i, move) {
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

	var count int
	for {
		var err error
		moves := randMoves(b)
		b, err = r.CreateNextBoardState(b, moves)
		if err != nil {
			return nil, err
		}

		count++

		over, err := r.IsGameOver(b)
		if err != nil {
			return nil, err
		}

		if count < 25 && !over {
			continue
		}

		res := make(map[int]float64)
		for i := 0; i < len(b.Snakes); i++ {
			if b.Snakes[i].EliminatedCause != "" {
				res[i] = -1
			} else if over {
				res[i] = 1
			} else {
				res[i] = 0
			}
		}
		return res, nil
	}
}

const C = 4.0

func selection(root *node, logd logd) *node {
	n := root
	for {
		if n.IsLeaf() {
			return n
		}

		type stats struct {
			sumN      float64
			sumTotals float64
		}

		allStats := make(map[int]map[string]stats)

		// Calculate stats for each move for each snake
		for _, tuple := range n.childs {
			if tuple.child.n == 0 {
				logd("select unexplored child, depth=%d, edge=%s", tuple.child.depth, tuple.edge)
				return tuple.child
			}

			for i := 0; i < len(n.idsByIdx); i++ {
				snakeStats, ok := allStats[i]
				if !ok {
					snakeStats = make(map[string]stats)
					allStats[i] = snakeStats
				}

				for _, move := range Moves {
					if !tuple.edge.Is(i, move) {
						continue
					}
					st := snakeStats[move]
					st.sumN += tuple.child.n
					st.sumTotals += tuple.child.totals[i]
					snakeStats[move] = st
					break
				}
			}
		}

		if n.n < 20 {
			random := n.childs[rand.Intn(len(n.childs))]
			n = random.child
			logd("select random child, depth=%d, edge=%s", n.depth, random.edge)
			continue
		}

		maxMoves := make(map[int]string)
		for i := 0; i < len(n.idsByIdx); i++ {
			max := float64(math.MinInt32)
			for _, move := range Moves {
				st := allStats[i][move]
				meanScore := st.sumTotals / st.sumN
				ucb1 := meanScore + math.Sqrt(C*math.Log(n.n)/st.sumN)
				if ucb1 > max {
					max = ucb1
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
			fmt.Printf("JCR: maxMoves=%v\n", maxMoves)
			fmt.Printf("JCR: n.childs=%v\n", n.childs)
			panic("missing child")
		}
		n = next
		logd("select DUCT child, depth=%d, edge=%s", n.depth, maxEdge)
	}
}

func propagation(node *node, totals map[int]float64) {
	n := node
	for n != nil {
		for idx, t := range totals {
			n.totals[idx] += t
		}
		n.n++
		n = n.parent
	}
}

func genMoveSet(board *rules.BoardState) []map[int]string {
	res := make([]map[int]string, 1)

	clone := func(m map[int]string) map[int]string {
		res := make(map[int]string)
		for k, v := range m {
			res[k] = v
		}
		return res
	}

	for i := 0; i < len(board.Snakes); i++ {
		if board.Snakes[i].EliminatedCause != "" {
			continue
		}

		var temp []map[int]string
		for mi, move := range Moves {
			if !isRationalMove(board, i, move) {
				// Skip unless it will result in 0 moves
				if len(temp) > 0 || mi < 3 {
					// fmt.Printf("JCR: skipping irrational move=%v, snake=%d\n", move, i)
					continue
				}
				// fmt.Printf("JCR: adding irrational move=%v, snake=%d\n", move, i)
			} // else {
			// fmt.Printf("JCR: adding ratinonal move=%v, snake=%d\n", move, i)
			//}

			for _, prev := range res {
				next := clone(prev)
				next[i] = move
				temp = append(temp, next)
			}
		}

		res = temp
	}
	return res
}

func isRationalMove(board *rules.BoardState, snakeIdx int, move string) bool {
	next := movePoint(board.Snakes[snakeIdx].Body[0], move)

	if next.X < 0 || next.X >= board.Width {
		return false
	}

	if next.Y < 0 || next.Y >= board.Height {
		return false
	}

	for i := 0; i < len(board.Snakes); i++ {
		for j := 0; j < len(board.Snakes[i].Body)-1; j++ {
			if next.X == board.Snakes[i].Body[j].X && next.Y == board.Snakes[i].Body[j].Y {
				return false
			}
		}
	}

	return true
}

func movePoint(p rules.Point, move string) rules.Point {
	switch move {
	case "up":
		return rules.Point{X: p.X, Y: p.Y + 1}
	case "down":
		return rules.Point{X: p.X, Y: p.Y - 1}
	case "left":
		return rules.Point{X: p.X - 1, Y: p.Y}
	case "right":
		return rules.Point{X: p.X + 1, Y: p.Y}
	}
	panic("unknown move")
}
