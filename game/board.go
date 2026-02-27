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
