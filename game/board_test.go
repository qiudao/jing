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
