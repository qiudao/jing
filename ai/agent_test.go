package ai

import (
	"testing"
)

func TestAgentGreedy(t *testing.T) {
	q := NewQTable()
	q.Set("s", 0, 0.1)
	q.Set("s", 4, 0.9)
	q.Set("s", 8, 0.5)
	agent := NewAgent(q, 0.0) // epsilon=0 -> always greedy
	action := agent.ChooseAction("s", []int{0, 4, 8})
	if action != 4 {
		t.Errorf("greedy agent should pick best action 4, got %d", action)
	}
}

func TestAgentFullRandom(t *testing.T) {
	q := NewQTable()
	agent := NewAgent(q, 1.0) // epsilon=1 -> always random
	counts := map[int]int{}
	for i := 0; i < 1000; i++ {
		a := agent.ChooseAction("s", []int{0, 1, 2})
		counts[a]++
	}
	for _, a := range []int{0, 1, 2} {
		if counts[a] < 200 {
			t.Errorf("action %d only chosen %d times, expected ~333", a, counts[a])
		}
	}
}
