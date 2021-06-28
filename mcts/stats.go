package mcts

import "math"

func graphDepths(root *node) []float64 {
	var res []float64

	q := []*node{root}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		res = append(res, float64(n.depth))
		for i := 0; i < len(n.childs); i++ {
			q = append(q, n.childs[i].child)
		}
	}
	return res
}

type sstats struct {
	count float64

	// Depths
	max    float64
	min    float64
	mean   float64
	stddev float64
	sum    float64
}

func sampleStats(samples []float64) sstats {
	var res sstats

	// Calculate count, sum and max
	for i, s := range samples {
		res.count++
		res.sum += s

		if i == 0 || res.max < s {
			res.max = s
		}
		if i == 0 || res.min > s {
			res.min = s
		}
	}

	// Calculate mean
	res.mean = res.sum / res.count

	// Calculate stddev
	var temp float64
	for _, s := range samples {
		temp += math.Pow(s-res.mean, 2)
	}
	res.stddev = math.Sqrt(temp / (res.count - 1))

	return res
}
