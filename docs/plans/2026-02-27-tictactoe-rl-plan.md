# Tic-Tac-Toe RL Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go web application where a Q-Learning AI learns to play Tic-Tac-Toe through self-play, with support for human vs human, human vs AI, and AI vs AI modes.

**Architecture:** Pure game logic layer (`game/`) -> RL layer (`ai/`) -> WebSocket server (`server/`) -> embedded static frontend (`web/`). All layers communicate through Go interfaces. Single binary with `embed.FS`.

**Tech Stack:** Go 1.25, gorilla/websocket, HTML/CSS/JS (no framework), Canvas for charts

---

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `main.go` (placeholder)

**Step 1: Initialize Go module**

Run: `cd /Users/k/work/jing && go mod init github.com/k/tictactoe-rl`
Expected: `go.mod` created

**Step 2: Create Makefile**

```makefile
.DEFAULT_GOAL := help

.PHONY: help build run test train clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o bin/tictactoe .

run: build ## Build and run the server
	./bin/tictactoe

test: ## Run all tests
	go test ./... -v

clean: ## Remove build artifacts and models
	rm -rf bin/

lint: ## Run go vet
	go vet ./...
```

**Step 3: Create placeholder main.go**

```go
package main

import "fmt"

func main() {
	fmt.Println("tictactoe-rl")
}
```

**Step 4: Verify build**

Run: `make build && ./bin/tictactoe`
Expected: prints `tictactoe-rl`

**Step 5: Commit**

```bash
git init
git add go.mod Makefile main.go docs/
git commit -m "init: project scaffold with Makefile and design docs"
```

---

### Task 2: Game Board Logic

**Files:**
- Create: `game/board.go`
- Create: `game/board_test.go`

**Step 1: Write failing tests for board**

```go
package game

import "testing"

func TestNewBoard(t *testing.T) {
	b := NewBoard()
	for i := 0; i < 9; i++ {
		if b.Cells[i] != Empty {
			t.Errorf("cell %d should be empty, got %d", i, b.Cells[i])
		}
	}
	if b.Turn != X {
		t.Errorf("first turn should be X, got %d", b.Turn)
	}
}

func TestMove(t *testing.T) {
	b := NewBoard()
	err := b.Move(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Cells[4] != X {
		t.Errorf("cell 4 should be X, got %d", b.Cells[4])
	}
	if b.Turn != O {
		t.Errorf("turn should be O after X moves, got %d", b.Turn)
	}
}

func TestMoveInvalid(t *testing.T) {
	b := NewBoard()
	b.Move(4)
	err := b.Move(4)
	if err == nil {
		t.Error("expected error for occupied cell")
	}
	err = b.Move(9)
	if err == nil {
		t.Error("expected error for out of range")
	}
}

func TestWinRows(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{1, 1, 1, 0, 0, 0, 0, 0, 0}
	if b.Winner() != X {
		t.Error("X should win with top row")
	}
}

func TestWinCols(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{2, 0, 0, 2, 0, 0, 2, 0, 0}
	if b.Winner() != O {
		t.Error("O should win with left column")
	}
}

func TestWinDiag(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{1, 0, 0, 0, 1, 0, 0, 0, 1}
	if b.Winner() != X {
		t.Error("X should win with main diagonal")
	}
}

func TestDraw(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{1, 2, 1, 1, 1, 2, 2, 1, 2}
	if b.Winner() != Empty {
		t.Error("should be no winner")
	}
	if !b.IsFull() {
		t.Error("board should be full")
	}
}

func TestAvailableMoves(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{1, 0, 0, 0, 2, 0, 0, 0, 1}
	moves := b.AvailableMoves()
	expected := []int{1, 2, 3, 5, 6, 7}
	if len(moves) != len(expected) {
		t.Fatalf("expected %d moves, got %d", len(expected), len(moves))
	}
	for i, m := range moves {
		if m != expected[i] {
			t.Errorf("move %d: expected %d, got %d", i, expected[i], m)
		}
	}
}

func TestStateKey(t *testing.T) {
	b := NewBoard()
	b.Cells = [9]int{1, 0, 2, 0, 1, 0, 0, 0, 2}
	key := b.StateKey()
	if key != "102010002" {
		t.Errorf("expected '102010002', got '%s'", key)
	}
}

func TestClone(t *testing.T) {
	b := NewBoard()
	b.Move(0)
	c := b.Clone()
	c.Move(4)
	if b.Cells[4] != Empty {
		t.Error("clone should not affect original")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/k/work/jing && go test ./game/ -v`
Expected: FAIL — types and functions not defined

**Step 3: Implement board.go**

```go
package game

import (
	"errors"
	"strconv"
	"strings"
)

const (
	Empty = 0
	X     = 1
	O     = 2
)

var winLines = [][3]int{
	{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, // rows
	{0, 3, 6}, {1, 4, 7}, {2, 5, 8}, // cols
	{0, 4, 8}, {2, 4, 6},             // diags
}

type Board struct {
	Cells [9]int
	Turn  int
}

func NewBoard() *Board {
	return &Board{Turn: X}
}

func (b *Board) Move(pos int) error {
	if pos < 0 || pos > 8 {
		return errors.New("position out of range")
	}
	if b.Cells[pos] != Empty {
		return errors.New("cell already occupied")
	}
	b.Cells[pos] = b.Turn
	if b.Turn == X {
		b.Turn = O
	} else {
		b.Turn = X
	}
	return nil
}

func (b *Board) Winner() int {
	for _, line := range winLines {
		a, bb, c := b.Cells[line[0]], b.Cells[line[1]], b.Cells[line[2]]
		if a != Empty && a == bb && bb == c {
			return a
		}
	}
	return Empty
}

func (b *Board) IsFull() bool {
	for _, c := range b.Cells {
		if c == Empty {
			return false
		}
	}
	return true
}

func (b *Board) IsOver() bool {
	return b.Winner() != Empty || b.IsFull()
}

func (b *Board) AvailableMoves() []int {
	moves := make([]int, 0, 9)
	for i, c := range b.Cells {
		if c == Empty {
			moves = append(moves, i)
		}
	}
	return moves
}

func (b *Board) StateKey() string {
	var sb strings.Builder
	for _, c := range b.Cells {
		sb.WriteString(strconv.Itoa(c))
	}
	return sb.String()
}

func (b *Board) Clone() *Board {
	return &Board{
		Cells: b.Cells,
		Turn:  b.Turn,
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./game/ -v`
Expected: all PASS

**Step 5: Commit**

```bash
git add game/
git commit -m "feat: add game board logic with full test coverage"
```

---

### Task 3: Q-Table Data Structure

**Files:**
- Create: `ai/qtable.go`
- Create: `ai/qtable_test.go`

**Step 1: Write failing tests**

```go
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
	// should return one of the available actions (first one since all are 0)
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./ai/ -v`
Expected: FAIL

**Step 3: Implement qtable.go**

```go
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
```

**Step 4: Run tests**

Run: `go test ./ai/ -v`
Expected: all PASS

**Step 5: Commit**

```bash
git add ai/
git commit -m "feat: add Q-Table data structure with thread-safe access"
```

---

### Task 4: AI Agent (Epsilon-Greedy)

**Files:**
- Create: `ai/agent.go`
- Create: `ai/agent_test.go`

**Step 1: Write failing tests**

```go
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
	// with 1000 tries and 3 options, each should get at least 200
	for _, a := range []int{0, 1, 2} {
		if counts[a] < 200 {
			t.Errorf("action %d only chosen %d times, expected ~333", a, counts[a])
		}
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./ai/ -v -run TestAgent`
Expected: FAIL

**Step 3: Implement agent.go**

```go
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
```

**Step 4: Run tests**

Run: `go test ./ai/ -v -run TestAgent`
Expected: all PASS

**Step 5: Commit**

```bash
git add ai/agent.go ai/agent_test.go
git commit -m "feat: add epsilon-greedy AI agent"
```

---

### Task 5: Trainer (Self-Play)

**Files:**
- Create: `ai/trainer.go`
- Create: `ai/trainer_test.go`

**Step 1: Write failing tests**

```go
package ai

import (
	"testing"

	"github.com/k/tictactoe-rl/game"
)

func TestTrainBasic(t *testing.T) {
	q := NewQTable()
	cfg := TrainConfig{
		Episodes: 1000,
		Alpha:    0.1,
		Gamma:    0.9,
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

	// trained agent (greedy) vs random agent, 1000 games
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
	// well-trained agent should lose < 5% to random
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./ai/ -v -run TestTrain`
Expected: FAIL

**Step 3: Implement trainer.go**

```go
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
	Episode  int
	Total    int
	XWins    int
	OWins    int
	Draws    int
	WinRateX float64 // rolling win rate for X in recent 100 games
}

type TrainResult struct {
	Episodes int
}

func Train(q *QTable, cfg TrainConfig, progress func(TrainProgress)) TrainResult {
	recentX, recentO, recentD := 0, 0, 0
	windowSize := 100
	results := make([]int, 0, windowSize) // 1=X, 2=O, 0=draw

	for ep := 0; ep < cfg.Episodes; ep++ {
		// linear epsilon decay
		epsilon := cfg.EpsilonStart - (cfg.EpsilonStart-cfg.EpsilonEnd)*float64(ep)/float64(cfg.Episodes)
		agentX := NewAgent(q, epsilon)
		agentO := NewAgent(q, epsilon)

		b := game.NewBoard()
		// record state-action history for both players
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

		// assign rewards and update Q-table
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
			nextState := ""
			var nextAvail []int
			if i+2 < len(history) {
				// next state from this player's perspective is 2 steps ahead
				nextState = history[i+2].state
				// reconstruct available moves at that point — approximate with current q entries
			}
			if nextState == "" {
				nextAvail = []int{}
			} else {
				// use terminal state, no future value
				nextAvail = []int{}
			}
			q.Update(s.state, s.action, reward, nextState, nextAvail, cfg.Alpha, cfg.Gamma)
		}

		// track results for rolling window
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
```

**Step 4: Run tests**

Run: `go test ./ai/ -v -run TestTrain -timeout 60s`
Expected: all PASS (TestTrainedAgentNeverLosesToRandom may take a few seconds)

**Step 5: Commit**

```bash
git add ai/trainer.go ai/trainer_test.go
git commit -m "feat: add self-play trainer with progress callback"
```

---

### Task 6: Model Persistence

**Files:**
- Create: `ai/model.go`
- Create: `ai/model_test.go`

**Step 1: Write failing tests**

```go
package ai

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoad(t *testing.T) {
	q := NewQTable()
	q.Set("s1", 0, 0.5)
	q.Set("s1", 4, 0.9)
	q.Set("s2", 2, -0.3)

	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")

	err := SaveQTable(q, path)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := LoadQTable(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Get("s1", 0) != 0.5 {
		t.Errorf("expected 0.5, got %f", loaded.Get("s1", 0))
	}
	if loaded.Get("s1", 4) != 0.9 {
		t.Errorf("expected 0.9, got %f", loaded.Get("s1", 4))
	}
	if loaded.Get("s2", 2) != -0.3 {
		t.Errorf("expected -0.3, got %f", loaded.Get("s2", 2))
	}
}

func TestLoadNonExistent(t *testing.T) {
	_, err := LoadQTable("/nonexistent/path.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestSaveCreatesDir(t *testing.T) {
	q := NewQTable()
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "model.json")
	err := SaveQTable(q, path)
	if err != nil {
		t.Fatalf("save should create parent dirs: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./ai/ -v -run TestSave`
Expected: FAIL

**Step 3: Implement model.go**

```go
package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func SaveQTable(q *QTable, path string) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.Marshal(q.Data)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadQTable(path string) (*QTable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	q := NewQTable()
	if err := json.Unmarshal(data, &q.Data); err != nil {
		return nil, err
	}
	return q, nil
}
```

**Step 4: Run tests**

Run: `go test ./ai/ -v -run "TestSave|TestLoad"`
Expected: all PASS

**Step 5: Commit**

```bash
git add ai/model.go ai/model_test.go
git commit -m "feat: add Q-Table JSON persistence"
```

---

### Task 7: WebSocket Server

**Files:**
- Modify: `go.mod` (add gorilla/websocket)
- Create: `server/handler.go`
- Create: `server/hub.go`

**Step 1: Add dependency**

Run: `cd /Users/k/work/jing && go get github.com/gorilla/websocket`

**Step 2: Implement hub.go**

```go
package server

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
}

func (h *Hub) Send(conn *websocket.Conn, msgType string, data interface{}) {
	raw, err := json.Marshal(data)
	if err != nil {
		log.Printf("marshal error: %v", err)
		return
	}
	msg := Message{Type: msgType, Data: raw}
	payload, _ := json.Marshal(msg)
	conn.WriteMessage(websocket.TextMessage, payload)
}
```

**Step 3: Implement handler.go**

This is the largest file. It handles all WebSocket message types and game state management.

```go
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/k/tictactoe-rl/ai"
	"github.com/k/tictactoe-rl/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type GameSession struct {
	Board      *game.Board
	Mode       string // hvh, hva, ava
	Difficulty string // beginner, intermediate, expert
	Speed      int    // ms per move for ava mode
	StopCh     chan struct{}
}

type Stats struct {
	mu   sync.Mutex
	Data map[string]*ModeStats
}

type ModeStats struct {
	XWins int `json:"xWins"`
	OWins int `json:"oWins"`
	Draws int `json:"draws"`
}

type Handler struct {
	hub       *Hub
	models    map[string]*ai.QTable
	modelDir  string
	sessions  map[*websocket.Conn]*GameSession
	mu        sync.Mutex
	stats     *Stats
	training  map[string]chan struct{} // difficulty -> stop channel
	trainMu   sync.Mutex
}

func NewHandler(modelDir string) *Handler {
	h := &Handler{
		hub:      NewHub(),
		models:   make(map[string]*ai.QTable),
		modelDir: modelDir,
		sessions: make(map[*websocket.Conn]*GameSession),
		stats: &Stats{
			Data: map[string]*ModeStats{
				"hvh": {},
				"hva-beginner": {},
				"hva-intermediate": {},
				"hva-expert": {},
				"ava": {},
			},
		},
		training: make(map[string]chan struct{}),
	}
	h.loadModels()
	return h
}

func (h *Handler) loadModels() {
	for _, diff := range []string{"beginner", "intermediate", "expert"} {
		path := filepath.Join(h.modelDir, diff+".json")
		q, err := ai.LoadQTable(path)
		if err != nil {
			log.Printf("no model for %s, will need training", diff)
			continue
		}
		h.models[diff] = q
		log.Printf("loaded model: %s (%d states)", diff, len(q.Data))
	}
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}
	h.hub.Add(conn)
	defer func() {
		h.mu.Lock()
		if sess, ok := h.sessions[conn]; ok {
			if sess.StopCh != nil {
				close(sess.StopCh)
			}
			delete(h.sessions, conn)
		}
		h.mu.Unlock()
		h.hub.Remove(conn)
		conn.Close()
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		h.handleMessage(conn, msg)
	}
}

func (h *Handler) handleMessage(conn *websocket.Conn, msg Message) {
	switch msg.Type {
	case "start_game":
		h.handleStartGame(conn, msg.Data)
	case "move":
		h.handleMove(conn, msg.Data)
	case "train":
		h.handleTrain(conn, msg.Data)
	case "get_stats":
		h.sendStats(conn)
	}
}

type StartGameData struct {
	Mode       string `json:"mode"`
	Difficulty string `json:"difficulty"`
	Speed      int    `json:"speed"`
}

type MoveData struct {
	Position int `json:"position"`
}

type TrainData struct {
	Difficulty string `json:"difficulty"`
	Watch      bool   `json:"watch"` // slow mode with visualization
	Speed      int    `json:"speed"` // ms per move in watch mode
}

func (h *Handler) handleStartGame(conn *websocket.Conn, data json.RawMessage) {
	var d StartGameData
	json.Unmarshal(data, &d)

	h.mu.Lock()
	if old, ok := h.sessions[conn]; ok && old.StopCh != nil {
		close(old.StopCh)
	}
	sess := &GameSession{
		Board:      game.NewBoard(),
		Mode:       d.Mode,
		Difficulty: d.Difficulty,
		Speed:      d.Speed,
	}
	h.sessions[conn] = sess
	h.mu.Unlock()

	h.sendGameState(conn, sess)

	if d.Mode == "ava" {
		sess.StopCh = make(chan struct{})
		go h.runAVA(conn, sess)
	}
}

func (h *Handler) handleMove(conn *websocket.Conn, data json.RawMessage) {
	var d MoveData
	json.Unmarshal(data, &d)

	h.mu.Lock()
	sess, ok := h.sessions[conn]
	h.mu.Unlock()
	if !ok {
		return
	}

	if err := sess.Board.Move(d.Position); err != nil {
		h.hub.Send(conn, "error", map[string]string{"message": err.Error()})
		return
	}

	if sess.Board.IsOver() {
		h.recordResult(sess)
		h.sendGameState(conn, sess)
		h.sendStats(conn)
		return
	}

	// AI response for hva mode
	if sess.Mode == "hva" {
		h.sendGameState(conn, sess)
		q, ok := h.models[sess.Difficulty]
		if !ok {
			h.hub.Send(conn, "error", map[string]string{"message": "model not trained yet"})
			return
		}
		epsilon := difficultyEpsilon(sess.Difficulty)
		agent := ai.NewAgent(q, epsilon)
		action := agent.ChooseAction(sess.Board.StateKey(), sess.Board.AvailableMoves())
		sess.Board.Move(action)
		if sess.Board.IsOver() {
			h.recordResult(sess)
		}
	}

	h.sendGameState(conn, sess)
	if sess.Board.IsOver() {
		h.sendStats(conn)
	}
}

func (h *Handler) runAVA(conn *websocket.Conn, sess *GameSession) {
	diff := sess.Difficulty
	if diff == "" {
		diff = "expert"
	}
	q, ok := h.models[diff]
	if !ok {
		h.hub.Send(conn, "error", map[string]string{"message": "model not trained yet"})
		return
	}
	epsilon := difficultyEpsilon(diff)
	agentX := ai.NewAgent(q, epsilon)
	agentO := ai.NewAgent(q, epsilon)
	speed := sess.Speed
	if speed < 100 {
		speed = 500
	}

	for !sess.Board.IsOver() {
		select {
		case <-sess.StopCh:
			return
		default:
		}
		var action int
		if sess.Board.Turn == game.X {
			action = agentX.ChooseAction(sess.Board.StateKey(), sess.Board.AvailableMoves())
		} else {
			action = agentO.ChooseAction(sess.Board.StateKey(), sess.Board.AvailableMoves())
		}
		sess.Board.Move(action)
		h.sendGameState(conn, sess)
		time.Sleep(time.Duration(speed) * time.Millisecond)
	}
	h.recordResult(sess)
	h.sendStats(conn)
}

func (h *Handler) handleTrain(conn *websocket.Conn, data json.RawMessage) {
	var d TrainData
	json.Unmarshal(data, &d)

	h.trainMu.Lock()
	if ch, ok := h.training[d.Difficulty]; ok {
		close(ch)
	}
	stopCh := make(chan struct{})
	h.training[d.Difficulty] = stopCh
	h.trainMu.Unlock()

	episodes := difficultyEpisodes(d.Difficulty)
	q := ai.NewQTable()

	go func() {
		if d.Watch {
			h.trainWithWatch(conn, q, d, episodes, stopCh)
		} else {
			cfg := ai.TrainConfig{
				Episodes:     episodes,
				Alpha:        0.1,
				Gamma:        0.9,
				EpsilonStart: 1.0,
				EpsilonEnd:   0.01,
			}
			ai.Train(q, cfg, func(p ai.TrainProgress) {
				select {
				case <-stopCh:
					return
				default:
				}
				h.hub.Send(conn, "train_progress", p)
			})
		}

		path := filepath.Join(h.modelDir, d.Difficulty+".json")
		ai.SaveQTable(q, path)

		h.mu.Lock()
		h.models[d.Difficulty] = q
		h.mu.Unlock()

		h.hub.Send(conn, "train_done", map[string]string{"difficulty": d.Difficulty})

		h.trainMu.Lock()
		delete(h.training, d.Difficulty)
		h.trainMu.Unlock()
	}()
}

func (h *Handler) trainWithWatch(conn *websocket.Conn, q *ai.QTable, d TrainData, episodes int, stopCh chan struct{}) {
	speed := d.Speed
	if speed < 100 {
		speed = 300
	}

	recentX, recentO, recentD := 0, 0, 0
	windowSize := 100
	results := make([]int, 0, windowSize)

	for ep := 0; ep < episodes; ep++ {
		select {
		case <-stopCh:
			return
		default:
		}

		epsilon := 1.0 - (1.0-0.01)*float64(ep)/float64(episodes)
		agentX := ai.NewAgent(q, epsilon)
		agentO := ai.NewAgent(q, epsilon)

		b := game.NewBoard()
		type step struct {
			state  string
			action int
			player int
		}
		var history []step

		// show this game to the client
		h.hub.Send(conn, "watch_game", map[string]interface{}{
			"episode": ep + 1,
			"total":   episodes,
		})

		for !b.IsOver() {
			select {
			case <-stopCh:
				return
			default:
			}
			state := b.StateKey()
			var action int
			if b.Turn == game.X {
				action = agentX.ChooseAction(state, b.AvailableMoves())
			} else {
				action = agentO.ChooseAction(state, b.AvailableMoves())
			}
			history = append(history, step{state, action, b.Turn})
			b.Move(action)

			h.hub.Send(conn, "game_state", map[string]interface{}{
				"board":  b.Cells,
				"turn":   b.Turn,
				"status": gameStatus(b),
				"winner": b.Winner(),
			})
			time.Sleep(time.Duration(speed) * time.Millisecond)
		}

		// update Q-table
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
			q.Update(s.state, s.action, reward, "", []int{}, 0.1, 0.9)
		}

		// rolling stats
		var result int
		switch winner {
		case game.X:
			result = 1
		case game.O:
			result = 2
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

		total := recentX + recentO + recentD
		winRate := 0.0
		if total > 0 {
			winRate = float64(recentX) / float64(total)
		}
		h.hub.Send(conn, "train_progress", ai.TrainProgress{
			Episode:  ep + 1,
			Total:    episodes,
			XWins:    recentX,
			OWins:    recentO,
			Draws:    recentD,
			WinRateX: winRate,
		})
	}
}

func (h *Handler) sendGameState(conn *websocket.Conn, sess *GameSession) {
	h.hub.Send(conn, "game_state", map[string]interface{}{
		"board":  sess.Board.Cells,
		"turn":   sess.Board.Turn,
		"status": gameStatus(sess.Board),
		"winner": sess.Board.Winner(),
	})
}

func gameStatus(b *game.Board) string {
	if b.Winner() != game.Empty {
		return fmt.Sprintf("win")
	}
	if b.IsFull() {
		return "draw"
	}
	return "playing"
}

func (h *Handler) recordResult(sess *GameSession) {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	key := sess.Mode
	if sess.Mode == "hva" {
		key = "hva-" + sess.Difficulty
	}
	s, ok := h.stats.Data[key]
	if !ok {
		s = &ModeStats{}
		h.stats.Data[key] = s
	}
	switch sess.Board.Winner() {
	case game.X:
		s.XWins++
	case game.O:
		s.OWins++
	default:
		s.Draws++
	}
}

func (h *Handler) sendStats(conn *websocket.Conn) {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()
	h.hub.Send(conn, "stats", h.stats.Data)
}

func difficultyEpsilon(d string) float64 {
	switch d {
	case "beginner":
		return 0.5
	case "intermediate":
		return 0.2
	case "expert":
		return 0.0
	}
	return 0.0
}

func difficultyEpisodes(d string) int {
	switch d {
	case "beginner":
		return 10000
	case "intermediate":
		return 50000
	case "expert":
		return 200000
	}
	return 50000
}
```

**Step 4: Verify build**

Run: `go build ./server/`
Expected: compiles without errors

**Step 5: Commit**

```bash
git add server/ go.mod go.sum
git commit -m "feat: add WebSocket server with game, training, and stats handlers"
```

---

### Task 8: Frontend (HTML/CSS/JS)

**Files:**
- Create: `web/index.html`
- Create: `web/style.css`
- Create: `web/app.js`

This is a single task but the files are long. The implementation details are in the design doc. Key points:

- `index.html`: semantic layout with three panels + training section
- `style.css`: CSS Grid for board, flexbox for layout, responsive
- `app.js`: WebSocket client, game state rendering, training controls, Canvas chart

**Step 1: Create web/index.html**

Complete HTML with all controls, board, stats panel, and training section. Use `embed` directive in Go to serve.

**Step 2: Create web/style.css**

Board as 3x3 CSS Grid with hover effects, responsive layout, progress bar styling, panel cards.

**Step 3: Create web/app.js**

WebSocket message handler, board click events, mode/difficulty switching, training trigger, Canvas line chart for convergence curve, speed slider for AVA mode.

**Step 4: Verify in browser**

Run the server, open http://localhost:8080, check all controls render.

**Step 5: Commit**

```bash
git add web/
git commit -m "feat: add web frontend with game board, controls, and training panel"
```

---

### Task 9: Main Entry Point & Wiring

**Files:**
- Modify: `main.go`

**Step 1: Wire everything together**

```go
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/k/tictactoe-rl/server"
)

//go:embed web/*
var webFS embed.FS

func main() {
	modelDir := "models"
	if d := os.Getenv("MODEL_DIR"); d != "" {
		modelDir = d
	}
	modelDir, _ = filepath.Abs(modelDir)

	addr := ":8080"
	if a := os.Getenv("ADDR"); a != "" {
		addr = a
	}

	handler := server.NewHandler(modelDir)

	webContent, _ := fs.Sub(webFS, "web")
	http.Handle("/", http.FileServer(http.FS(webContent)))
	http.HandleFunc("/ws", handler.ServeWS)

	log.Printf("starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
```

**Step 2: Build and run**

Run: `make run`
Expected: server starts on :8080

**Step 3: Smoke test in browser**

Open http://localhost:8080, verify page loads, WebSocket connects.

**Step 4: Commit**

```bash
git add main.go
git commit -m "feat: wire up main entry point with embedded frontend"
```

---

### Task 10: Integration Test & Polish

**Step 1: Run all unit tests**

Run: `make test`
Expected: all PASS

**Step 2: Manual test all modes**

- Train beginner model
- Play human vs AI (beginner)
- Play human vs human
- Watch AI vs AI
- Check stats update
- Train with watch mode

**Step 3: Fix any issues found**

**Step 4: Final commit**

```bash
git add -A
git commit -m "polish: integration fixes and final cleanup"
```
