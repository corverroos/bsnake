package main

import (
	"context"
	"fmt"
	"math"
	"time"
)

func selectMove(ctx context.Context, req GameRequest, w weights) (string, error) {
	max := math.MinInt64
	var res Move
	for _, m := range Moves {
		score, err := scoreMove(ctx, req, w, m,false)
		if err != nil {
			return "", err
		}
		if score > max {
			max = score
			res = m
		}
	}

	if max == math.MinInt64 {
		return "", fmt.Errorf("no safe moves")
	}
	return res.String(), nil
}


type weights struct {
	Wall int
	Back int
	Body int
	Tail int
	MyTail int
	Hole int
	H2H int
	H2HWin int
	FoodFull int
	HungryHealth int
}

func scoreMove(ctx context.Context, req GameRequest,  w weights, m Move, debug bool) (int, error) {
	t0 := time.Now()
	logd := func(msg string, args ...interface{}) {
		if !debug {
			return
		}
		suffix := fmt.Sprintf(msg, args...)
		fmt.Printf("%4d %s: %s\n", time.Since(t0).Milliseconds(),  string(m.String()[0]), suffix)
	}

	next := req.You.Head.Move(m)
	if isWall(req, next) {
		logd("wall")
		return w.Wall, nil
	} else if len(req.You.Body) > 1 && req.You.Body[1].Equal(next) {
		logd("no back")
		return w.Back, nil
	}

	score := 100

	if ttl, you, ok := isBody(req, next); ok && you && ttl == 1 {
		logd("my tail")
		score += w.MyTail
	} else if ok && ttl == 1 {
		logd("tail")
		score += w.Tail
	} else if ok {
		logd("body")
		return w.Body, nil
	}

	logd("filling %d", score)

	area := fill(req, req.You.Length*2, next)
	fill := area.Size()
	logd("size %d", score)

	isHole := fill < req.You.Length
	if isHole {
		logd("fill small hole")
		score += w.Hole
		if c, ttl := area.MinTTL(); ttl != 0 {
			dist := distance(next, c)
			score += dist - ttl
			logd("dist %d to ttl ttl %d, score %d", dist, ttl, score)
		}
	} else {
		score += fill
		logd("ok fill %d score %d", fill, score)
	}

	if win, ok := isPossibleH2H(req, next); ok && win {
		score += w.H2HWin
		logd("possible winning h2h %v", score)
	} else if ok && !win {
		score += w.H2H
		logd("possible loosing h2h %v", score)
	}

	hungry := isHungry(req, w)
	logd("hungry %v %v", hungry, score)

	if dist := closestFood(req, next); !isHole && !hungry && dist == 1 {
		score += w.FoodFull
		logd("food full dist %v %v", dist, score)
	} else if dist > 0 && hungry {
		score += 50 / dist
		logd("food hungry dist %v %v", dist, score)
	}
	logd("score %v", score)
	return score, nil
}
