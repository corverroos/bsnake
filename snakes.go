package main

import "context"

type snake struct {
	Info BattlesnakeInfoResponse
	Start func(ctx context.Context, req GameRequest) error
	End func(ctx context.Context,req GameRequest) error
	Move func(ctx context.Context,req GameRequest) (string, error)
}

var snakes = map[string]snake{
	"basic" : {
		Info:  BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "spaceworm",
			Color:      "#000000",
			Head:       "sand-worm",
			Tail:       "round-bum",
		},
		Start: nil,
		End:   nil,
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			return selectMove(ctx, req, basicWeights)
		},
	},
	"ball" : {
		Info:  BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "spaceball",
			Color:      "#00264d",
			Head:       "bendr",
			Tail:       "freckled",
		},
		Start: nil,
		End:   nil,
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			return selectMove(ctx, req, ballWeights)
		},
	},
	"monty" : {
		Info:  BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "spacetree",
		},
		Start: nil,
		End:   nil,
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			return selectMCTS(ctx, req, basicWeights)
		},
	},
}


const (
	scoreWall = -1001
	scoreBack = -1002
	scoreBody = -1003
)

var basicWeights = weights{
	Wall:         scoreWall,
	Back:         scoreBack,
	Body:         scoreBody,
	MyTail: +10,
	Tail:         -50,
	Hole:         -100,
	H2H:          -100,
	H2HWin:       +20,
	FoodFull:     -20,
	HungryHealth: 30,
}

var ballWeights = weights{
	Wall:         scoreWall,
	Back:         scoreBack,
	Body:         scoreBody,
	MyTail:         10,
	Hole:         -100,
	H2H:          -100,
	H2HWin:       -100,
	FoodFull:     -100,
	HungryHealth: 10,
}
