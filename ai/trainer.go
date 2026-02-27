package ai

import "github.com/k/tictactoe-rl/game"

type TrainConfig struct {
	Episodes     int
	Alpha        float64
	Gamma        float64
	EpsilonStart float64
	EpsilonEnd   float64
}

type TrainProgress struct {
	Episode  int     `json:"episode"`
	Total    int     `json:"total"`
	XWins    int     `json:"xWins"`
	OWins    int     `json:"oWins"`
	Draws    int     `json:"draws"`
	WinRateX float64 `json:"winRateX"`
}

type TrainResult struct {
	Episodes int
}

func Train(q *QTable, cfg TrainConfig, progress func(TrainProgress)) TrainResult {
	recentX, recentO, recentD := 0, 0, 0
	windowSize := 100
	results := make([]int, 0, windowSize)

	for ep := 0; ep < cfg.Episodes; ep++ {
		epsilon := cfg.EpsilonStart - (cfg.EpsilonStart-cfg.EpsilonEnd)*float64(ep)/float64(cfg.Episodes)
		agentX := NewAgent(q, epsilon)
		agentO := NewAgent(q, epsilon)

		b := game.NewBoard()
		type step struct {
			state     string
			action    int
			player    int
			available []int
		}
		var history []step

		for !b.IsOver() {
			state := b.StateKey()
			avail := b.AvailableMoves()
			var action int
			if b.Turn == game.X {
				action = agentX.ChooseAction(state, avail)
			} else {
				action = agentO.ChooseAction(state, avail)
			}
			history = append(history, step{state, action, b.Turn, avail})
			b.Move(action)
		}

		winner := b.Winner()
		for i := len(history) - 1; i >= 0; i-- {
			s := history[i]

			// find next state from this player's perspective (2 steps ahead)
			var nextState string
			var nextAvail []int
			if i+2 < len(history) {
				nextState = history[i+2].state
				nextAvail = history[i+2].available
			}

			// only terminal steps get explicit reward
			var reward float64
			if i+2 >= len(history) {
				// this is one of the last two moves — terminal
				switch {
				case winner == game.Empty:
					reward = 0.5
				case winner == s.player:
					reward = 1.0
				default:
					reward = -1.0
				}
			}
			// non-terminal steps: reward=0, value comes from gamma*maxQ(nextState)

			q.Update(s.state, s.action, reward, nextState, nextAvail, cfg.Alpha, cfg.Gamma)
		}

		var result int
		switch winner {
		case game.X:
			result = 1
		case game.O:
			result = 2
		default:
			result = 0
		}

		if len(results) >= windowSize {
			old := results[0]
			results = results[1:]
			switch old {
			case 1:
				recentX--
			case 2:
				recentO--
			default:
				recentD--
			}
		}
		results = append(results, result)
		switch result {
		case 1:
			recentX++
		case 2:
			recentO++
		default:
			recentD++
		}

		if progress != nil && (ep%100 == 0 || ep == cfg.Episodes-1) {
			total := recentX + recentO + recentD
			winRate := 0.0
			if total > 0 {
				winRate = float64(recentX) / float64(total)
			}
			progress(TrainProgress{
				Episode:  ep + 1,
				Total:    cfg.Episodes,
				XWins:    recentX,
				OWins:    recentO,
				Draws:    recentD,
				WinRateX: winRate,
			})
		}
	}

	return TrainResult{Episodes: cfg.Episodes}
}
