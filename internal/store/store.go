package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Command struct {
	ID          int      `json:"id"`
	Command     string   `json:"command"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

type Store struct {
	path string
}

func New() (*Store, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(dir, "favs", "commands.json")}, nil
}

func NewAt(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Load() ([]Command, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []Command{}, nil
	}
	if err != nil {
		return nil, err
	}
	var commands []Command
	if err := json.Unmarshal(data, &commands); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", s.path, err)
	}
	return commands, nil
}

func (s *Store) Save(commands []Command) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(commands, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}

func (s *Store) Add(command string, tags []string, description string) (Command, error) {
	commands, err := s.Load()
	if err != nil {
		return Command{}, err
	}
	cmd := Command{
		ID:          nextID(commands),
		Command:     command,
		Tags:        tags,
		Description: description,
	}
	return cmd, s.Save(append(commands, cmd))
}

func (s *Store) Remove(id int) error {
	commands, err := s.Load()
	if err != nil {
		return err
	}
	for i, cmd := range commands {
		if cmd.ID == id {
			return s.Save(append(commands[:i], commands[i+1:]...))
		}
	}
	return fmt.Errorf("no command with id %d", id)
}

func (s *Store) Filter(tag string) ([]Command, error) {
	commands, err := s.Load()
	if err != nil {
		return nil, err
	}
	if tag == "" {
		return commands, nil
	}
	var filtered []Command
	for _, cmd := range commands {
		for _, t := range cmd.Tags {
			if t == tag {
				filtered = append(filtered, cmd)
				break
			}
		}
	}
	return filtered, nil
}

func nextID(commands []Command) int {
	max := 0
	for _, cmd := range commands {
		if cmd.ID > max {
			max = cmd.ID
		}
	}
	return max + 1
}
