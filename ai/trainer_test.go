package ai

import (
	"testing"

	"github.com/k/tictactoe-rl/game"
)

func TestTrainBasic(t *testing.T) {
	q := NewQTable()
	cfg := TrainConfig{
		Episodes:     1000,
		Alpha:        0.1,
		Gamma:        0.9,
		EpsilonStart: 1.0,
		EpsilonEnd:   0.01,
	}
	result := Train(q, cfg, nil)
	if result.Episodes != 1000 {
		t.Errorf("expected 1000 episodes, got %d", result.Episodes)
	}
	if len(q.Data) == 0 {
		t.Error("q-table should have entries after training")
	}
}

func TestTrainedAgentNeverLosesToRandom(t *testing.T) {
	q := NewQTable()
	cfg := TrainConfig{
		Episodes:     50000,
		Alpha:        0.1,
		Gamma:        0.9,
		EpsilonStart: 1.0,
		EpsilonEnd:   0.01,
	}
	Train(q, cfg, nil)

	trained := NewAgent(q, 0.0)
	random := NewAgent(NewQTable(), 1.0)
	losses := 0
	for i := 0; i < 1000; i++ {
		b := game.NewBoard()
		for !b.IsOver() {
			var action int
			if b.Turn == game.X {
				action = trained.ChooseAction(b.StateKey(), b.AvailableMoves())
			} else {
				action = random.ChooseAction(b.StateKey(), b.AvailableMoves())
			}
			b.Move(action)
		}
		if b.Winner() == game.O {
			losses++
		}
	}
	if losses > 50 {
		t.Errorf("trained agent lost %d/1000 games to random, too many", losses)
	}
}

func TestTrainProgress(t *testing.T) {
	q := NewQTable()
	cfg := TrainConfig{
		Episodes:     100,
		Alpha:        0.1,
		Gamma:        0.9,
		EpsilonStart: 1.0,
		EpsilonEnd:   0.01,
	}
	var updates []TrainProgress
	callback := func(p TrainProgress) {
		updates = append(updates, p)
	}
	Train(q, cfg, callback)
	if len(updates) == 0 {
		t.Error("should have received progress updates")
	}
}
