package server

import (
	"encoding/json"
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
	Mode       string
	Difficulty string
	Speed      int
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
	hub      *Hub
	models   map[string]*ai.QTable
	modelDir string
	sessions map[*websocket.Conn]*GameSession
	mu       sync.Mutex
	stats    *Stats
	training map[string]chan struct{}
	trainMu  sync.Mutex
}

func NewHandler(modelDir string) *Handler {
	h := &Handler{
		hub:      NewHub(),
		models:   make(map[string]*ai.QTable),
		modelDir: modelDir,
		sessions: make(map[*websocket.Conn]*GameSession),
		stats: &Stats{
			Data: map[string]*ModeStats{
				"hvh":              {},
				"hva-beginner":     {},
				"hva-intermediate": {},
				"hva-expert":       {},
				"ava":              {},
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
	Watch      bool   `json:"watch"`
	Speed      int    `json:"speed"`
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

	// Check model exists before starting AI modes
	if d.Mode == "hva" || d.Mode == "ava" {
		if _, ok := h.models[d.Difficulty]; !ok {
			h.hub.Send(conn, "error", map[string]string{
				"message": "model not trained yet for " + d.Difficulty + ", please train first",
			})
			h.mu.Lock()
			delete(h.sessions, conn)
			h.mu.Unlock()
			return
		}
	}

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
			state     string
			action    int
			player    int
			available []int
		}
		var history []step

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
			avail := b.AvailableMoves()
			var action int
			if b.Turn == game.X {
				action = agentX.ChooseAction(state, avail)
			} else {
				action = agentO.ChooseAction(state, avail)
			}
			history = append(history, step{state, action, b.Turn, avail})
			b.Move(action)

			h.hub.Send(conn, "game_state", map[string]interface{}{
				"board":  b.Cells,
				"turn":   b.Turn,
				"status": gameStatus(b),
				"winner": b.Winner(),
			})
			time.Sleep(time.Duration(speed) * time.Millisecond)
		}

		winner := b.Winner()
		for i := len(history) - 1; i >= 0; i-- {
			s := history[i]
			var nextState string
			var nextAvail []int
			if i+2 < len(history) {
				nextState = history[i+2].state
				nextAvail = history[i+2].available
			}
			var reward float64
			if i+2 >= len(history) {
				switch {
				case winner == game.Empty:
					reward = 0.5
				case winner == s.player:
					reward = 1.0
				default:
					reward = -1.0
				}
			}
			q.Update(s.state, s.action, reward, nextState, nextAvail, 0.1, 0.9)
		}

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
		return "win"
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
