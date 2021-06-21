package mcts

import (
	"math"
	"math/rand"
	"sort"

	"github.com/BattlesnakeOfficial/rules"
)

var Moves = []string{"up", "down", "right", "left"}

func RandMoves() []string {
	var res []string
	for _, move := range Moves {
		res = append(res, move)
	}
	rand.Shuffle(len(res), func(i, j int) {
		res[i], res[j] = res[j], res[i]
	})
	return res
}

type edge int32

func (e edge) Is(idx int, move string) bool {
	switch 0b111 & (e >> (idx * 3)) {
	case 0:
		return false
	case 1:
		return move == "up"
	case 2:
		return move == "down"
	case 3:
		return move == "left"
	case 4:
		return move == "right"
	default:
		panic("invalid")
	}
}

func (e edge) String() string {
	i := int32(e)
	var res string
	for i > 0 {
		switch 0b111 & i {
		case 0:
			res += "_"
		case 1:
			res += "u"
		case 2:
			res += "d"
		case 3:
			res += "l"
		case 4:
			res += "r"
		default:
			panic("invalid")
		}
		i = i >> 3
	}
	return res
}

func newEdge(moves map[int]string) edge {
	var v int32
	for idx, move := range moves {
		switch move {
		case "up":
			v |= 1 << (idx * 3)
		case "down":
			v |= 2 << (idx * 3)
		case "left":
			v |= 3 << (idx * 3)
		case "right":
			v |= 4 << (idx * 3)
		default:
			panic("invalid")
		}
	}
	return edge(v)
}

type tuple struct {
	edge  edge
	child *node
}

type node struct {
	ruleset  rules.Ruleset
	idsByIdx map[int]string
	rootIdx  int

	board *rules.BoardState
	depth int

	parent    *node
	childs    []tuple
	lastMoves map[int]string

	n      float64
	totals map[int]float64

	termTotals map[int]float64
}

func (n *node) MinMaxMove(idx int) string {
	mins := make(map[string]float64)
	for _, tuple := range n.childs {
		for _, move := range Moves {
			if tuple.edge.Is(idx, move) {
				avg := tuple.child.AvgScore(idx)
				min, ok := mins[move]
				if !ok || min > avg {
					mins[move] = avg
				}
			}
		}
	}

	var (
		res string
		max = float64(math.MinInt32)
	)

	for move, min := range mins {
		if min > max {
			res = move
			max = min
		}
	}

	return res
}

func (n *node) MinAvgScore(idx int, move string) float64 {
	var min = float64(math.MaxInt32)
	for _, tuple := range n.childs {
		if tuple.edge.Is(idx, move) {
			avg := tuple.child.AvgScore(idx)
			if min > avg {
				min = avg
			}
		}
	}
	return min
}

func (n *node) RobustSafeMove(idx int) string {
	var first string
	for i, move := range n.RobustMoves(idx) {
		if i == 0 {
			first = move
		}
		if n.MinAvgScore(idx, move) <= -1 {
			continue
		}
		return move
	}

	return first
}

func (n *node) RobustMoves(idx int) []string {
	type tup struct {
		K string
		V float64
	}

	totals := make([]tup, len(Moves))

	for _, tuple := range n.childs {
		for i, move := range Moves {
			if tuple.edge.Is(idx, move) {
				totals[i].K = move
				totals[i].V += tuple.child.n
			}
		}
	}

	sort.Slice(totals, func(i, j int) bool {
		return totals[i].V > totals[j].V
	})

	var res []string
	for _, t := range totals {
		if t.K == "" {
			continue
		}
		res = append(res, t.K)
	}
	return res
}

func (n *node) UpdateScores(s map[int]float64) {
	if n.totals == nil {
		n.totals = make(map[int]float64)
	}
	n.n++
	for i, t := range s {
		n.totals[i] += t
	}
}

func (n *node) IsLeaf() bool {
	return len(n.childs) == 0
}

func (n *node) IsTerminal() bool {
	return len(n.termTotals) > 0
}

func (n *node) CheckTerminal() (map[int]float64, bool, error) {
	if n.board.Snakes[n.rootIdx].EliminatedCause == "" {
		// We are still alive
		if ok, err := n.ruleset.IsGameOver(n.board); err != nil {
			return nil, false, err
		} else if !ok {
			// Game not over
			return nil, false, nil
		}
	}

	// We are dead or game is over

	res := make(map[int]float64)
	for idx, snake := range n.board.Snakes {
		score := -1.0
		if snake.EliminatedCause == "" {
			score = 1
		}
		res[idx] = score
	}

	return res, true, nil
}

func (n *node) AvgScore(snakeIDx int) float64 {
	if n.n == 0 {
		return 0
	}
	return n.totals[snakeIDx] / n.n
}

func (n *node) UCB1(snakeIDx int) (float64, bool) {
	if n.n == 0 {
		return 0, true
	}

	return n.AvgScore(snakeIDx) + math.Sqrt(2)*math.Sqrt(math.Log(n.parent.n)/n.n), false
}

func (n *node) GenChild(moves map[int]string) (*node, error) {
	if _, ok := moves[n.rootIdx]; !ok {
		panic("missing root idx")
	}

	e := newEdge(moves)

	var ml []rules.SnakeMove
	for idx, move := range moves {
		ml = append(ml, rules.SnakeMove{
			ID:   n.idsByIdx[idx],
			Move: move,
		})
	}

	board, err := n.ruleset.CreateNextBoardState(n.board, ml)
	if err != nil {
		return nil, err
	}

	child := &node{
		ruleset:   n.ruleset,
		idsByIdx:  n.idsByIdx,
		rootIdx:   n.rootIdx,
		board:     board,
		depth:     n.depth + 1,
		lastMoves: moves,
		parent:    n,
		totals:    make(map[int]float64),
	}
	n.childs = append(n.childs, tuple{edge: e, child: child})

	return child, nil
}

func NewRoot(ruleset rules.Ruleset, board *rules.BoardState, rootIdx int) *node {
	idsByIdx := make(map[int]string)
	for idx, snake := range board.Snakes {
		idsByIdx[idx] = snake.ID
	}
	return &node{
		ruleset:  ruleset,
		idsByIdx: idsByIdx,
		rootIdx:  rootIdx,
		board:    board,
		n:        1,
		totals:   make(map[int]float64),
	}
}
