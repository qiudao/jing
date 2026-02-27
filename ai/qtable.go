package ai

import "sync"

type QTable struct {
	mu   sync.RWMutex
	Data map[string]map[int]float64 `json:"data"`
}

func NewQTable() *QTable {
	return &QTable{
		Data: make(map[string]map[int]float64),
	}
}

func (q *QTable) Get(state string, action int) float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if actions, ok := q.Data[state]; ok {
		return actions[action]
	}
	return 0.0
}

func (q *QTable) Set(state string, action int, value float64) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.Data[state]; !ok {
		q.Data[state] = make(map[int]float64)
	}
	q.Data[state][action] = value
}

func (q *QTable) BestAction(state string, available []int) int {
	best := available[0]
	bestVal := q.Get(state, best)
	for _, a := range available[1:] {
		v := q.Get(state, a)
		if v > bestVal {
			bestVal = v
			best = a
		}
	}
	return best
}

func (q *QTable) MaxQ(state string, available []int) float64 {
	if len(available) == 0 {
		return 0.0
	}
	maxVal := q.Get(state, available[0])
	for _, a := range available[1:] {
		v := q.Get(state, a)
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

func (q *QTable) Update(state string, action int, reward float64, nextState string, nextAvailable []int, alpha, gamma float64) {
	oldQ := q.Get(state, action)
	maxNext := q.MaxQ(nextState, nextAvailable)
	newQ := oldQ + alpha*(reward+gamma*maxNext-oldQ)
	q.Set(state, action, newQ)
}
