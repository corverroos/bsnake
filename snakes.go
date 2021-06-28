package main

import (
	"context"

	"github.com/BattlesnakeOfficial/rules"

	"github.com/corverroos/bsnake/heur"
	"github.com/corverroos/bsnake/mcts"
)

type snake struct {
	Alias       string
	Description string
	Info        BattlesnakeInfoResponse
	Start       func(ctx context.Context, req GameRequest) error
	End         func(ctx context.Context, req GameRequest) error
	Move        func(ctx context.Context, req GameRequest) (string, error)
}

var snakes = map[string]snake{
	"v0": {
		Description: "First snake, just weighted next move heuristic",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "corverroos",
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
	"mx0": {
		Description: "Minimax 2-ply",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Meta:       fmx0,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMinimax(board, coordsToPoints(req.Board.Hazards), rootIdx, fmx0, mxDepth(board))
		},
	},
	"mx1": {
		Description: "Minimax 4-ply",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Meta:       fmx1,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMinimax(board, coordsToPoints(req.Board.Hazards), rootIdx, fmx1, mxDepth(board))
		},
	},
	"mx2": {
		Description: "Minimax 4-ply",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Meta:       fmx2,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMinimax(board, coordsToPoints(req.Board.Hazards), rootIdx, fmx2, mxDepth(board))
		},
	},
	"mx3": {
		Description: "Minimax 4-ply",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Meta:       fmx3,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMinimax(board, coordsToPoints(req.Board.Hazards), rootIdx, fmx3, mxDepth(board))
		},
	},
	"mx4": {
		Description: "Minimax Tree Search",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Color: "#ccccff",
			Head: "snow-worm",
			Tail: "block-bum",
			Meta: map[string]interface{}{
				"heur": fmx3,
				"opts": mcts.OptsV2,
			},
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMx(board, coordsToPoints(req.Board.Hazards), rootIdx, fmx3, &mcts.OptsV2)
		},
	},
	"v1": {
		Description: "MCTS with multiplayer, simultaneous move, Decoupled-UCT, rational-playouts",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "corverroos",
			Color:      "#141452",
			Head:       "viper",
			Tail:       "rattle",
			Meta:       mcts.OptsV1,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMove(ctx, board, coordsToPoints(req.Board.Hazards), rootIdx, &mcts.OptsV1)
		},
	},
	"v2": {
		Description: "MCTS with multiplayer, simultaneous move, Decoupled-UCT, rational-playouts",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "corverroos",
			Color:      "#ff0066",
			Head:       "silly",
			Tail:       "rocket",
			Meta:       mcts.OptsV2,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMove(ctx, board, coordsToPoints(req.Board.Hazards), rootIdx, &mcts.OptsV2)
		},
	},
	"v3": {
		Alias:       "latest",
		Description: "MCTS with multiplayer, simultaneous move, Decoupled-UCT, rational-playouts",
		Info: BattlesnakeInfoResponse{
			APIVersion: "1",
			Author:     "corverroos",
			Color:      "#141452",
			Head:       "viper",
			Tail:       "rattle",
			Meta:       mcts.OptsV3,
		},
		Move: func(ctx context.Context, req GameRequest) (string, error) {
			board, rootIdx := gameReqToBoard(req)
			return mcts.SelectMove(ctx, board, coordsToPoints(req.Board.Hazards), rootIdx, &mcts.OptsV3)
		},
	},
}

var (
	fmx0 = heur.Factors{
		Control: 0.01,
		Length:  0.5,
		Hunger:  -0.001,
		Starve:  -0.9,
	}
	fmx1 = heur.Factors{
		Control: 0.1,
		Length:  0.3,
		Hunger:  -0.02,
		Starve:  -0.8,
	}
	fmx2 = heur.Factors{
		Control: 0.1,
		Length:  0.3,
		Hunger:  -0.001,
		Starve:  -0.9,
	}
	fmx3 = heur.Factors{
		Control: 0.05,
		Length:  0.35,
		Hunger:  -0.001,
		Starve:  -0.9,
	}
)

func init() {
	for _, s := range snakes {
		if s.Alias != "" {
			snakes[s.Alias] = s
		}
	}
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
	MyTail:       +10,
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
	MyTail:       10,
	Hole:         -100,
	H2H:          -100,
	H2HWin:       -100,
	FoodFull:     -100,
	HungryHealth: 10,
}

func mxDepth(b *rules.BoardState) int {
	var l int
	for i := 0; i < len(b.Snakes); i++ {
		if b.Snakes[i].EliminatedCause == "" {
			l++
		}
	}
	switch l {
	case 1:
		return 10
	case 2:
		return 4
	case 3:
		return 2
	case 4:
		return 2
	default:
		return 1
	}
}
