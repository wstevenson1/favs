package picker_test

import (
	"errors"
	"testing"

	"github.com/wstevenson1/favs/internal/picker"
)

func TestParseSelection_ValidNumber(t *testing.T) {
	n, err := picker.ParseSelection("2", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2, got %d", n)
	}
}

func TestParseSelection_QuitLowercase(t *testing.T) {
	_, err := picker.ParseSelection("q", 5)
	if !errors.Is(err, picker.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestParseSelection_QuitUppercase(t *testing.T) {
	_, err := picker.ParseSelection("Q", 5)
	if !errors.Is(err, picker.ErrQuit) {
		t.Errorf("expected ErrQuit, got %v", err)
	}
}

func TestParseSelection_OutOfRange(t *testing.T) {
	_, err := picker.ParseSelection("6", 5)
	if !errors.Is(err, picker.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestParseSelection_ZeroIsInvalid(t *testing.T) {
	_, err := picker.ParseSelection("0", 5)
	if !errors.Is(err, picker.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestParseSelection_NonNumeric(t *testing.T) {
	_, err := picker.ParseSelection("abc", 5)
	if !errors.Is(err, picker.ErrInvalid) {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}

func TestParseSelection_WhitespaceTrimmed(t *testing.T) {
	n, err := picker.ParseSelection("  3  ", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Errorf("expected 3, got %d", n)
	}
}
