package store_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wstevenson1/favs/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	return store.NewAt(filepath.Join(t.TempDir(), "commands.toml"))
}

func TestLoad_EmptyWhenFileAbsent(t *testing.T) {
	s := newTestStore(t)
	cmds, err := s.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 0 {
		t.Fatalf("expected empty slice, got %v", cmds)
	}
}

func TestAdd_AssignsMonotonicIDs(t *testing.T) {
	s := newTestStore(t)
	c1, err := s.Add("echo hello", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := s.Add("echo world", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if c1.ID != 1 {
		t.Errorf("expected id 1, got %d", c1.ID)
	}
	if c2.ID != 2 {
		t.Errorf("expected id 2, got %d", c2.ID)
	}
}

func TestAdd_PersistsAcrossLoads(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Add("echo hello", []string{"test"}, "test cmd"); err != nil {
		t.Fatal(err)
	}
	cmds, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Command != "echo hello" {
		t.Errorf("expected 'echo hello', got %q", cmds[0].Command)
	}
}

func TestRemove_RemovesByID(t *testing.T) {
	s := newTestStore(t)
	c, _ := s.Add("echo hello", nil, "")
	if err := s.Remove(c.ID); err != nil {
		t.Fatal(err)
	}
	cmds, _ := s.Load()
	if len(cmds) != 0 {
		t.Errorf("expected empty list after remove, got %v", cmds)
	}
}

func TestRemove_ErrorOnUnknownID(t *testing.T) {
	s := newTestStore(t)
	err := s.Remove(99)
	if err == nil {
		t.Fatal("expected error for unknown id, got nil")
	}
}

func TestFilter_ByTag(t *testing.T) {
	s := newTestStore(t)
	s.Add("echo hello", []string{"foo"}, "")
	s.Add("echo world", []string{"bar"}, "")
	cmds, err := s.Filter("foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 1 || cmds[0].Command != "echo hello" {
		t.Errorf("unexpected filter result: %v", cmds)
	}
}

func TestFilter_EmptyTagReturnsAll(t *testing.T) {
	s := newTestStore(t)
	s.Add("echo hello", []string{"foo"}, "")
	s.Add("echo world", []string{"bar"}, "")
	cmds, err := s.Filter("")
	if err != nil {
		t.Fatal(err)
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 commands, got %d", len(cmds))
	}
}

func TestSave_WritesTOMLFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commands.toml")
	s := store.NewAt(path)
	if _, err := s.Add("echo hello", []string{"foo", "bar"}, "test desc"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[[command]]") {
		t.Errorf("expected TOML array-of-tables syntax, got:\n%s", content)
	}
	if !strings.Contains(content, `command = "echo hello"`) {
		t.Errorf("expected command field, got:\n%s", content)
	}
	if !strings.Contains(content, "tags = [") {
		t.Errorf("expected tags array, got:\n%s", content)
	}
}
