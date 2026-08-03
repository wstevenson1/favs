package picker_test

import (
	"testing"

	"github.com/wstevenson1/favs/internal/picker"
)

func TestPick_EmptyListReturnsEmptyStringWithoutError(t *testing.T) {
	selected, err := picker.Pick(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if selected != "" {
		t.Errorf("expected empty selection, got %q", selected)
	}
}
