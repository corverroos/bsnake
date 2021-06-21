package mcts

import "github.com/BattlesnakeOfficial/rules"

type RoyaleRuleset struct {
	rules.StandardRuleset
	Hazards []rules.Point
}

func (r *RoyaleRuleset) CreateNextBoardState(prevState *rules.BoardState,
	moves []rules.SnakeMove) (*rules.BoardState, error) {

	nextBoardState, err := r.StandardRuleset.CreateNextBoardState(prevState, moves)
	if err != nil {
		return nil, err
	}

	r.damageOutOfBounds(nextBoardState)

	return nextBoardState, nil
}

func (r *RoyaleRuleset) damageOutOfBounds(b *rules.BoardState) {
	for i := 0; i < len(b.Snakes); i++ {
		snake := &b.Snakes[i]
		if snake.EliminatedCause == "" {
			head := snake.Body[0]
			for _, p := range r.Hazards {
				if head == p {
					// Snake is now out of bounds, reduce health
					snake.Health = snake.Health - 15
					if snake.Health <= 0 {
						snake.Health = 0
						snake.EliminatedCause = "out-of-health"
					}
				}
			}
		}
	}
}
