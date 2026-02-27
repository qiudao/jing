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
