package heur

import (
	"sort"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/board"
)

type Factors struct {
	Control float64
	Length  float64
	Hunger  float64
	Starve  float64
}

func Calc(f *Factors, b *rules.BoardState, hazards map[rules.Point]bool) map[int]float64 {
	lens := Length(b)
	hunger := Hunger(b, hazards)

	var control map[int]float64
	var starve map[int]bool
	if f.Control > 0 {
		control, starve = Flood(b, hazards)
	}

	normalize(lens)
	normalize(control)
	normalize(hunger)

	res := make(map[int]float64)

	for i, v := range control {
		res[i] += f.Control * v
	}

	for i, v := range lens {
		res[i] += f.Length * v
	}

	for i, v := range hunger {
		res[i] += f.Hunger * v
	}

	for k := range res {
		if starve[k] {
			res[k] += f.Starve
		}
	}

	return res
}

func Hunger(b *rules.BoardState, hazards map[rules.Point]bool) map[int]float64 {
	minFood := make(map[int]float64)

	for i := 0; i < len(b.Snakes); i++ {
		s := &b.Snakes[i]
		if s.EliminatedCause != "" {
			continue
		}
		for _, point := range b.Food {
			dist := float64(board.Distance(s.Body[0], point))
			if hazards[point] {
				dist *= 2
			}
			if prev, ok := minFood[i]; !ok || prev > dist {
				minFood[i] = dist
			}
		}
	}

	return minFood
}

func Flood(b *rules.BoardState, hazards map[rules.Point]bool) (map[int]float64, map[int]bool) {
	control := make(map[int]float64, len(b.Snakes))
	starve := make(map[int]bool, len(b.Snakes))
	visited := make(map[rules.Point]int, b.Height*b.Width)

	food := make(map[rules.Point]bool, len(b.Food))
	for _, point := range b.Food {
		food[point] = true
	}

	type E struct {
		Idx    int
		P      rules.Point
		Health int32
		Depth  int
	}

	q := make([]E, 0, b.Height*b.Width)
	for i := 0; i < len(b.Snakes); i++ {
		s := &b.Snakes[i]
		if s.EliminatedCause != "" {
			continue
		}
		q = append(q, E{Idx: i, P: s.Body[0], Health: s.Health})

		l := len(s.Body)
		for i, point := range s.Body {
			if i == 0 {
				visited[point] = 1
			} else {
				visited[point] = i - l
			}
		}
	}

	sort.Slice(q, func(i, j int) bool {
		return len(b.Snakes[q[i].Idx].Body) > len(b.Snakes[q[j].Idx].Body)
	})

	for len(q) > 0 {
		e := q[0]
		q = q[1:]

		control[e.Idx]++

		for i := 0; i < 4; i++ {
			next := e.P
			switch i {
			case 0:
				next.X += 1
			case 1:
				next.X -= 1
			case 2:
				next.Y += 1
			case 3:
				next.Y -= 1
			}

			if next.X == -1 || next.Y == -1 || next.X == b.Width || next.Y == b.Height {
				continue
			} else if prev, ok := visited[next]; ok && (prev > 0 || -prev > e.Depth) {
				continue
			}

			h := e.Health - 1
			if hazards[next] {
				h -= 15
			}

			if food[next] {
				starve[e.Idx] = false
				h = 100
			}

			if h <= 0 {
				if _, ok2 := starve[e.Idx]; !ok2 {
					starve[e.Idx] = true
				}
				continue
			}

			q = append(q, E{
				Idx:    e.Idx,
				P:      next,
				Health: h,
				Depth:  e.Depth + 1,
			})

			visited[next] = 1
		}
	}

	return control, starve
}

func normalize(in map[int]float64) {
	var total float64
	for _, v := range in {
		total += v
	}

	if total == 0 {
		return
	}

	l := float64(len(in))
	for k, v := range in {
		in[k] = (v - total/l) / total
	}
}

func Length(b *rules.BoardState) map[int]float64 {
	res := make(map[int]float64)
	for i := 0; i < len(b.Snakes); i++ {
		if b.Snakes[i].EliminatedCause != "" {
			continue
		}
		res[i] = float64(len(b.Snakes[i].Body))
	}

	return res
}

func SelectMove(f *Factors, b *rules.BoardState, hazards map[rules.Point]bool, rootIdx int) (string, error) {

	var maxHeur float64
	var maxMove string

	s := &b.Snakes[rootIdx]
	oldBody := s.Body[:len(s.Body)-1]

	for i, move := range []string{"up", "down", "left", "right"} {

		next := oldBody[0]
		switch i {
		case 0:
			next.Y += 1
		case 1:
			next.Y -= 1
		case 2:
			next.X -= 1
		case 3:
			next.X += 1
		}

		if next.X == -1 || next.Y == -1 || next.X == b.Width || next.Y == b.Height {
			continue
		}

		var collide bool
		for _, s := range b.Snakes {
			for j, b := range s.Body {
				if j == len(s.Body)-1 {
					break
				}
				if next.X == b.X && next.Y == b.Y {
					collide = true
					break
				}
			}
		}
		if collide {
			continue
		}

		s.Body = append([]rules.Point{next}, oldBody...)

		res := Calc(f, b, hazards)

		if maxMove == "" || maxHeur < res[rootIdx] {
			maxMove = move
			maxHeur = res[rootIdx]
		}
	}

	if maxMove == "" {
		return "UP", nil
	}

	return maxMove, nil
}
