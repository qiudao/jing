package ai

import "math/rand"

type Agent struct {
	QTable  *QTable
	Epsilon float64
}

func NewAgent(q *QTable, epsilon float64) *Agent {
	return &Agent{QTable: q, Epsilon: epsilon}
}

func (a *Agent) ChooseAction(state string, available []int) int {
	if rand.Float64() < a.Epsilon {
		return available[rand.Intn(len(available))]
	}
	return a.QTable.BestAction(state, available)
}
