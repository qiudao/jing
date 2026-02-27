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
			state  string
			action int
			player int
		}
		var history []step

		for !b.IsOver() {
			state := b.StateKey()
			var action int
			if b.Turn == game.X {
				action = agentX.ChooseAction(state, b.AvailableMoves())
			} else {
				action = agentO.ChooseAction(state, b.AvailableMoves())
			}
			history = append(history, step{state, action, b.Turn})
			b.Move(action)
		}

		winner := b.Winner()
		for i := len(history) - 1; i >= 0; i-- {
			s := history[i]
			var reward float64
			switch {
			case winner == game.Empty:
				reward = 0.5
			case winner == s.player:
				reward = 1.0
			default:
				reward = -1.0
			}
			q.Update(s.state, s.action, reward, "", []int{}, cfg.Alpha, cfg.Gamma)
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
