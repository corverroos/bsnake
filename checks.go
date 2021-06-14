package main

import (
	"math"
)


func isPossibleH2H(req GameRequest, c Coord) (win, ok bool) {
	for _, s := range req.Board.Snakes {
		if s.ID == req.You.ID {
			continue
		}
		if c.Equal(s.Head) {
			return false, true
		}
		for _, m := range Moves {
			if c.Equal(s.Head.Move(m)) {
				return req.You.Length > s.Length + 1, true
			}
		}
	}
	return false, false
}

func isWall(req GameRequest, c Coord) bool {
	if c.X >= req.Board.Width || c.X < 0 {
		return true
	}
	if c.Y >= req.Board.Height || c.Y < 0 {
		return true
	}

	return false
}

// isBody returns true if the coordinate is on a snake body and
// then returns the minimum ttl (turns-to-live).
func isBody(req GameRequest, c Coord) (int, bool, bool) {
	for _, s := range req.Board.Snakes {
		for i, b := range s.Body {
			if c.Equal(b) {
				return len(s.Body) - i, req.You.ID == s.ID, true
			}
		}
	}

	return 0, false, false
}

// fill returns the area from root with positive numbers for fills and negatives-ttls for bodies and 0 for walls.
func fill(req GameRequest, maxDepth int, root Coord) Area {
	type E struct {
		depth int
		c Coord
	}
	result := make(Area)
	q := []E{{depth: 1, c: root}}
	for len(q) > 0 {
		e := q[0]
		q = q[1:]

		c := e.c
		depth := e.depth

		if prev, ok := result[c]; ok && prev <= 0 {
			// Visited before and not safe, so always skip
			continue
		} else if ok && prev <= depth {
			// Visited again with higher depth, so skip
			continue
		} else if ok /* prev > depth */ {
			// Visiting again with lower depth, so continue below
		} else if !ok {
			// Not visited before, check for walls
			if isWall(req, c) {
				result[c] = 0
				continue
			}
			// Check for body
			if ttl, _, ok := isBody(req, c); ok {
				result[c] = -ttl
				continue
			}
			// Continue below
		} else {
			panic("bug")
		}

		result[c] = depth

		if maxDepth == depth {
			continue
		}

		for _, m := range Moves {
			next := c.Move(m)
			q = append(q, E{c: next, depth: depth+1})
		}
	}

	return result
}

func closestFood(req GameRequest, c Coord) int {
	if len(req.Board.Food) == 0 {
		return 0
	}

	min := math.MaxInt64
	for _, food := range req.Board.Food {
		dist := distance(food, c)
		if dist < min {
			min = dist
		}
	}

	return min+1
}

func distance(c1 Coord, c2 Coord) int {
	x := c1.X-c2.X
	y := c1.Y-c2.Y
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	return x + y
}


func isHungry(req GameRequest, w weights ) bool {
	if req.You.Health < w.HungryHealth {
		return true
	}
	for _, snake := range req.Board.Snakes {
		if snake.ID == req.You.ID {
			continue
		}
		if snake.Length >= req.You.Length - 2 {
			return true
		}
	}
	return false
}
