package main

import (
	"context"
	"fmt"
	"github.com/BattlesnakeOfficial/rules"
	"math"
	"reflect"
	"strings"
	"time"
)

var mctsC = 2.0 * 100.0 * math.Sqrt(2.0)

type mcnode struct {
	Board *rules.BoardState // Result of moves below applied to parent board state.
	n float64 // Number of visits
	t float64 // Total score
	dead bool
	turn int


	parent *mcnode
	move string // my previous move (suffix of upstream edge)
	childs map[edge]*mcnode // Child node by next own move after previous all moves.
}

type edge string

func newEdge(moves []rules.SnakeMove, next string) edge {
	if len(moves) == 0 {
		return edge(next)
	}

	var res strings.Builder
	for _, move := range moves {
		res.WriteByte(move.Move[0])
	}

		res.WriteString("_")
		res.WriteByte(next[0])

	return edge(res.String())
}

func (n *mcnode) IsLeaf() bool {
	return len(n.childs) == 0
}

func (n *mcnode) AvgScore() float64 {
	if n.n == 0 {
		return 0
	}
	return n.t / n.n
}

func (n *mcnode) UCB1() (float64, bool) {
	if n.n == 0 {
		return 0, true
	}
	if n.dead {
		return -100, false
	}
	return n.t/n.n + mctsC * math.Sqrt(math.Log(n.parent.n)/n.n), false
}

// AddChild adds a child for the given boardstate, the moves that resulted in it, and your next possible move.
func (n *mcnode) AddChild(b *rules.BoardState, moves []rules.SnakeMove, next string) *mcnode {
	if n.childs == nil {
		n.childs = make(map[edge]*mcnode)
	}

	edge := newEdge(moves, next)
	if _, ok := n.childs[edge]; ok {
		panic("bug: over writing edge")
	}

	child := &mcnode{
		Board:  b,
		turn:   n.turn+1,
		parent: n,
		move:   next,
	}
	n.childs[edge] = child
	return child
}

func MCTSOnce(ruleset rules.Ruleset, root *mcnode, youIDx int, logd func(string,...interface{})) error {
	if root.dead {
		return nil
	}
	node := root
	for !node.IsLeaf(){
		// Select node
		max := float64(math.MinInt64)
		var next *mcnode
		for _, n := range node.childs {
			if n.dead {
				continue
			}
			score, inf := n.UCB1()
			if inf {
				next = n
				break
			} else if score > max{
				next = n
				max = score
			}
		}
		node = next
		logd("select node turn=%d move=%v\n", node.turn, node.move)
	}

	if node.n > 0 {
		logd("expand\n")
		// Expand other moves
		var child *mcnode
		for _, moves := range genMoves(node.Board, node.move, youIDx) {
			board, err := ruleset.CreateNextBoardState(node.Board, moves)
			if err != nil {
				return err
			}

			for _, next := range Moves {
				if isDeadlyRulesMove(board, youIDx, next.String()) {
					continue
				}
				child = node.AddChild(board, moves, next.String())
			}
		}
		if child == nil {
			logd("expanded nothing, will die\n")
			node.dead = true
		} else {
			logd("expanded %d, picked xxxx_%s\n", len(node.childs), child.move)
			logd("node.childs=%#v\n", node.childs)
			node = child
		}

	}

	var score float64
	if isDeadlyRulesMove(node.Board, youIDx, node.move) {
		logd("dead\n")
		node.dead = true
	}

	if !node.dead{
		// Rollout
		logd("rollout\n")
		var err error
		score, err = RollOut(node.Board, ruleset, youIDx, node.move, logd)
		if err != nil {
			return err
		}
	} else {
		score = -100
	}

	// Propagate
	logd("Propagate score=%v\n", score)
	for {
		node.n++
		node.t += score

		if !node.IsLeaf() {
			allDead := true
			for _, n := range node.childs {
				if !n.dead {
					allDead = false
					break
				}
			}
			if allDead {
				logd("all children dead\n")
				node.dead = true
			}
		}

		if node.parent == nil {
			break
		} else {
			node = node.parent
		}
	}

	return nil
}

func genMoves(board *rules.BoardState, youMove string, youIDx int) [][]rules.SnakeMove {
	res := [][]rules.SnakeMove{{{ID:board.Snakes[youIDx].ID, Move: youMove}}}
	for i := 0; i < len(board.Snakes); i++ {
		if i == youIDx || board.Snakes[i].EliminatedCause != "" {
			continue
		}
		var temp [][]rules.SnakeMove
		for _, move := range Moves {
			if len(temp) > 0 && isDeadlyRulesMove(board, i, move.String()) {
				continue
			}
			for _, prev := range res {
				next := append([]rules.SnakeMove{{ID: board.Snakes[i].ID, Move: move.String()}}, prev...)
				temp = append(temp, next)
			}
		}
		res = temp
	}
	return res
}

func RollOut(in *rules.BoardState, ruleset rules.Ruleset, youIDx int, youFirst string, logd func(string,...interface{})) (float64, error) {
	b := in

	randMoves := func(b *rules.BoardState) []rules.SnakeMove {

		res := make([]rules.SnakeMove, len(b.Snakes))
		for i := 0; i < len(b.Snakes); i++ {
			for j, move := range RandMoves() {
				if j < 3 && isDeadlyRulesMove(b, i,  move.String()) {
					continue
				}
				res[i] = rules.SnakeMove{
					ID:   b.Snakes[i].ID,
					Move: move.String(),
				}
				break
			}
		}
		if youFirst != "" {
			res[youIDx].Move = youFirst
			youFirst = ""
		}
		return res
	}

	var rolls int
	for {
		var err error
		moves := randMoves(b)
		b, err = ruleset.CreateNextBoardState(b, moves)
		if err != nil {
			return 0, err
		}

		if b.Snakes[youIDx].EliminatedCause != "" {
			logd("rollout: dead %s", b.Snakes[youIDx].EliminatedCause)
			return -100, nil
		}

		if reflect.TypeOf(ruleset) == solorules && len(b.Snakes[youIDx].Body) == int(b.Height*b.Width) {
			//Filled the board
			logd("rollout: solo max")
			return 100, nil
		}

		if ok, err := ruleset.IsGameOver(b); err != nil {
			return 0, err
		} else if ok {
			logd("rollout: game over (win)\n")
			return 100, nil
		}
		
		rolls++
		if rolls > 10 {
			var killed float64
			for i := 0; i < len(b.Snakes); i++ {
				if i == youIDx {
					continue
				}
				if b.Snakes[i].EliminatedBy == b.Snakes[youIDx].ID {
					killed++
				}
			}
			suffix := ""
			if killed > 0 {
				suffix = fmt.Sprintf("%.0f",killed)
			}
			logd("rollout: end %s\n", suffix)
			return killed, nil
		}
	}
}

var solorules = reflect.TypeOf(&rules.SoloRuleset{})

func isDeadlyRulesMove(board *rules.BoardState, snakeIDx int, move string) bool {
	next := moveRule(board.Snakes[snakeIDx].Body[0],move)

	if next.X < 0 || next.X >= board.Width {
		return true
	}

	if next.Y < 0 || next.Y >= board.Height {
		return true
	}

	for i := 0; i < len(board.Snakes); i++ {
		for j := 0; j < len(board.Snakes[i].Body) - 1; j++ {
			if next.X == board.Snakes[i].Body[j].X && next.Y == board.Snakes[i].Body[j].Y {
				return true
			}
		}
	}

	return false
}

func moveRule(p rules.Point, move string) rules.Point {
	switch move {
	case "up":
		return rules.Point{X: p.X, Y: p.Y +1 }
	case "down":
		return rules.Point{X: p.X, Y: p.Y -1 }
	case "left":
		return rules.Point{X: p.X -1, Y: p.Y}
	case "right":
		return rules.Point{X: p.X +1, Y: p.Y}
	}
	panic("unknown move")
}


func gameReqToBoard(req GameRequest) (*rules.BoardState, int) {
	var snakes []rules.Snake
	var youIDx int
	for i, snake := range req.Board.Snakes {
		id := snake.Name
		if id == "" {
			id = snake.ID
		}
		snakes = append(snakes, rules.Snake{
			ID:              id,
			Body:            coordsToPoints(snake.Body),
			Health:          int32(snake.Health),
		})
		if req.You.ID == snake.ID {
			youIDx=i
		}
	}

	return &rules.BoardState{
		Height: int32(req.Board.Height),
		Width:  int32(req.Board.Width),
		Food:   coordsToPoints(req.Board.Food),
		Snakes: snakes,
	}, youIDx
}


func coordsToPoints(cl []Coord) (res []rules.Point) {
	for _, c := range cl {
		res = append(res, rules.Point{
			X: int32(c.X),
			Y: int32(c.Y),
		})
	}
	return res
}


func selectMCTS(ctx context.Context, req GameRequest, w weights) (string, error) {
	t0 := time.Now()

	m, err := selectMove(ctx, req, w)
	if err != nil {
		return "", err
	}

	board, youIDx := gameReqToBoard(req)
	root := &mcnode{
		Board:  board,
	}
	for _, move := range Moves {
		root.AddChild(board, nil, move.String())
	}

	ruleset := rules.Ruleset(&rules.StandardRuleset{})
	if len(req.Board.Snakes) == 1 {
		ruleset = &rules.SoloRuleset{}
	}

	for {
		err := MCTSOnce(ruleset, root, youIDx, func(string, ...interface{}) {})
		if err != nil {
			return "", err
		}

		if root.dead || time.Since(t0) > time.Millisecond * 100 {
			var choose *mcnode
			for _, node := range root.childs {
				if choose == nil || node.AvgScore() > choose.AvgScore() {
					choose = node
				}
			}

			if choose.AvgScore() == -100 {
				fmt.Printf("tree all -100, fallback to weights %s", m)
				return m, nil
			}

			go func(t0 time.Time) {
				logMCTS(t0, root, choose)
			}(t0)

			return choose.move, nil
		}
	}
}

func logMCTS(t0 time.Time, root, choose *mcnode) {
	var log strings.Builder
	log.WriteString(fmt.Sprintf("choose tree %s=%f\n", choose.move, choose.AvgScore()))
	log.WriteString("  other moves: ")
	for m, c := range root.childs {
		star := ""
		if c.dead {
			star = "*"
		}
		score, inf := c.UCB1()
		log.WriteString(fmt.Sprintf("%s=%.0f[%.0f%v] %v:%v ", m, c.AvgScore(), c.n, star,score, inf ))
	}

	log.WriteString(fmt.Sprintf("\n  graph: size=%d, duration=%s\n", gsize(root), time.Since(t0)))
	fmt.Print(log.String())
}

func gsize(n *mcnode) int {
	size := len(n.childs)
	for _, c := range n.childs {
		size += gsize(c)
	}
	return size
}
