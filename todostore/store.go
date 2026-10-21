package todostore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const StorageKey = "todo_taro_todos"

type Item struct {
	Id   int    `json:"Id"`
	Text string `json:"Text"`
	Done bool   `json:"Done"`
}

type Store struct {
	path   string
	items  []Item
	nextID int
}

func New(dataDir string) *Store {
	path := filepath.Join(dataDir, StorageKey+".json")
	s := &Store{path: path}
	s.load()
	return s
}

func (s *Store) load() {
	s.items = nil
	s.nextID = 1
	raw, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return
	}
	var arr []Item
	if err := json.Unmarshal(raw, &arr); err != nil {
		return
	}
	s.items = arr
	s.recomputeNextID()
}

func (s *Store) recomputeNextID() {
	s.nextID = 1
	for _, it := range s.items {
		if it.Id >= s.nextID {
			s.nextID = it.Id + 1
		}
	}
}

func (s *Store) save() {
	_ = os.MkdirAll(filepath.Dir(s.path), 0o755)
	b, _ := json.MarshalIndent(s.items, "", "  ")
	_ = os.WriteFile(s.path, b, 0o644)
}

func (s *Store) Items() []Item {
	return append([]Item(nil), s.items...)
}

func (s *Store) Add(text string) {
	t := strings.TrimSpace(text)
	if t == "" {
		return
	}
	s.items = append(s.items, Item{Id: s.nextID, Text: t, Done: false})
	s.nextID++
	s.save()
}

func (s *Store) addDone(text string) {
	t := strings.TrimSpace(text)
	if t == "" {
		return
	}
	s.items = append(s.items, Item{Id: s.nextID, Text: t, Done: true})
	s.nextID++
	s.save()
}

func (s *Store) SetDone(index int, done bool) {
	if index < 0 || index >= len(s.items) || s.items[index].Done == done {
		return
	}
	s.items[index].Done = done
	s.save()
}

func (s *Store) RemoveAt(index int) {
	if index < 0 || index >= len(s.items) {
		return
	}
	s.items = append(s.items[:index], s.items[index+1:]...)
	s.save()
}

func (s *Store) MoveItem(from, to int) {
	if from < 0 || from >= len(s.items) || to < 0 || to >= len(s.items) || from == to {
		return
	}
	row := s.items[from]
	slice := append(s.items[:from], s.items[from+1:]...)
	if to > from {
		to--
	}
	s.items = append(slice[:to], append([]Item{row}, slice[to:]...)...)
	s.save()
}

func (s *Store) ReplaceFromJSON(raw string) error {
	var arr []Item
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return err
	}
	s.items = arr
	s.recomputeNextID()
	s.save()
	return nil
}

func (s *Store) UpdateText(index int, text string) {
	t := strings.TrimSpace(text)
	if t == "" || index < 0 || index >= len(s.items) {
		return
	}
	s.items[index].Text = t
	s.save()
}

func (s *Store) ImportMarkdown(content string) {
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "- [x]") {
			s.addDone(strings.TrimSpace(line[5:]))
		} else if strings.HasPrefix(line, "- [ ]") {
			s.Add(strings.TrimSpace(line[5:]))
		} else {
			s.Add(line)
		}
	}
}

func (s *Store) ExportMarkdown() string {
	var lines []string
	for _, it := range s.items {
		if it.Done {
			lines = append(lines, "- [x] "+it.Text)
		} else {
			lines = append(lines, "- [ ] "+it.Text)
		}
	}
	return strings.Join(lines, "\n")
}

func (s *Store) CountsLabel() string {
	left, done := 0, 0
	for _, it := range s.items {
		if it.Done {
			done++
		} else {
			left++
		}
	}
	return fmt.Sprintf("还剩 %d · 已完成 %d", left, done)
}
