package heur

import (
	"math"
	"sort"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/board"
)

type Factors struct {
	Control float64
	Boxed   float64
	Length  float64
	Walls   float64
	Hunger  float64
	Starve  float64
}

func Calc(f *Factors, b *rules.BoardState, hazards map[rules.Point]bool) map[int]float64 {
	l := len(b.Snakes)

	res := make(map[int]float64, l)

	if f.Length != 0 {
		lengths := Length(b)
		normalize(lengths)
		for i := 0; i < l; i++ {
			res[i] += f.Length * lengths[i]
		}
	}

	if f.Hunger != 0 {
		hunger := Hunger(b, hazards)
		normalize(hunger)
		for i := 0; i < l; i++ {
			res[i] += f.Hunger * hunger[i]
		}
	}

	if f.Control != 0 || f.Starve != 0 || f.Boxed != 0 {
		control, starve := Flood(b, hazards)

		if f.Boxed != 0 {
			for i := 0; i < l; i++ {
				boxed := 1.0 - math.Min(1.0, control[i]/float64(len(b.Snakes[i].Body)))
				res[i] += f.Boxed * boxed
			}
		}

		normalize(control)

		for i := 0; i < l; i++ {
			res[i] += f.Control * control[i]

			if starve[i] == 1 {
				res[i] += f.Starve
			}
		}
	}

	if f.Walls != 0 {
		walls := Walls(b)
		for i := 0; i < l; i++ {
			res[i] += f.Walls * walls[i]
		}
	}

	for i := 0; i < l; i++ {
		if b.Snakes[i].EliminatedCause != "" {
			res[i] = -1
		}
	}

	return res
}

func Walls(b *rules.BoardState) []float64 {
	walls := make([]float64, len(b.Snakes))
	for i := 0; i < len(b.Snakes); i++ {
		if b.Snakes[i].EliminatedCause != "" {
			continue
		}

		h := b.Snakes[i].Body[0]
		w := b.Width - h.X
		if t := h.X + 1; t < w {
			w = t
		}
		if t := b.Height - h.Y; t < w {
			w = t
		}
		if t := h.Y + 1; t < w {
			w = t
		}
		walls[i] = float64(w) / float64(b.Height)
	}

	return walls
}

func Hunger(b *rules.BoardState, hazards map[rules.Point]bool) []float64 {
	minFood := make([]float64, len(b.Snakes))

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
			if prev := minFood[i]; prev == 0 || prev > dist {
				minFood[i] = dist
			}
		}
	}

	return minFood
}

func Flood(b *rules.BoardState, hazards map[rules.Point]bool) ([]float64, []int) {
	control := make([]float64, len(b.Snakes))
	starve := make([]int, len(b.Snakes)) // 1 == true, 0 or -1 == false

	visited := make([]int, b.Height*b.Width)
	food := make([]bool, b.Height*b.Width)

	pidx := func(p rules.Point) int {
		return int(p.Y*b.Width + p.X)
	}

	for _, point := range b.Food {
		food[pidx(point)] = true
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
		for i := 0; i < l; i++ {
			if i == 0 {
				visited[pidx(s.Body[i])] = 1
			} else {
				visited[pidx(s.Body[i])] = i - l
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

			nidx := pidx(next)

			if next.X == -1 || next.Y == -1 || next.X == b.Width || next.Y == b.Height {
				continue
			} else if prev := visited[nidx]; prev > 0 || -prev > e.Depth {
				continue
			}

			h := e.Health - 1
			if hazards[next] {
				h -= 15
			}

			if food[nidx] {
				starve[e.Idx] = -1
				h = 100
			}

			if h <= 0 {
				if starve[e.Idx] == 0 {
					starve[e.Idx] = 1
				}
				continue
			}

			q = append(q, E{
				Idx:    e.Idx,
				P:      next,
				Health: h,
				Depth:  e.Depth + 1,
			})

			visited[nidx] = 1
		}
	}

	return control, starve
}

func normalize(in []float64) {
	l := len(in)

	var total float64
	for i := 0; i < l; i++ {
		total += in[i]
	}

	if total == 0 {
		return
	}

	for i := 0; i < l; i++ {
		in[i] = (in[i] - total/float64(l)) / total
	}
}

func Length(b *rules.BoardState) []float64 {
	res := make([]float64, len(b.Snakes))
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
