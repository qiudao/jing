package ai

import "testing"

func TestNewQTable(t *testing.T) {
	q := NewQTable()
	if len(q.Data) != 0 {
		t.Error("new qtable should be empty")
	}
}

func TestGetQ(t *testing.T) {
	q := NewQTable()
	v := q.Get("state1", 3)
	if v != 0.0 {
		t.Errorf("unknown state-action should return 0, got %f", v)
	}
}

func TestSetAndGetQ(t *testing.T) {
	q := NewQTable()
	q.Set("state1", 3, 0.75)
	v := q.Get("state1", 3)
	if v != 0.75 {
		t.Errorf("expected 0.75, got %f", v)
	}
}

func TestBestAction(t *testing.T) {
	q := NewQTable()
	q.Set("s", 0, 0.1)
	q.Set("s", 1, 0.9)
	q.Set("s", 2, 0.5)
	action := q.BestAction("s", []int{0, 1, 2})
	if action != 1 {
		t.Errorf("expected action 1, got %d", action)
	}
}

func TestBestActionUnknownState(t *testing.T) {
	q := NewQTable()
	action := q.BestAction("unknown", []int{3, 5, 7})
	if action != 3 {
		t.Errorf("expected first available action 3, got %d", action)
	}
}

func TestUpdate(t *testing.T) {
	q := NewQTable()
	// Q(s,a) = Q(s,a) + alpha * (reward + gamma * maxQ(s') - Q(s,a))
	// Q = 0 + 0.1 * (1.0 + 0.9 * 0.0 - 0.0) = 0.1
	q.Update("s", 0, 1.0, "s2", []int{0, 1}, 0.1, 0.9)
	v := q.Get("s", 0)
	if v < 0.099 || v > 0.101 {
		t.Errorf("expected ~0.1, got %f", v)
	}
}
