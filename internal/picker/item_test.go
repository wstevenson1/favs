package picker

import (
	"strings"
	"testing"

	"github.com/wstevenson1/favs/internal/store"
)

func TestCommandItem_TitleWithTags(t *testing.T) {
	item := newCommandItem(store.Command{Command: "git log --oneline", Tags: []string{"git", "log"}})
	want := "[git,log] git log --oneline"
	if got := item.Title(); got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestCommandItem_TitleWithoutTags(t *testing.T) {
	item := newCommandItem(store.Command{Command: "df -h"})
	want := "df -h"
	if got := item.Title(); got != want {
		t.Errorf("Title() = %q, want %q", got, want)
	}
}

func TestCommandItem_Description(t *testing.T) {
	item := newCommandItem(store.Command{Command: "df -h", Description: "Disk usage"})
	if got := item.Description(); got != "Disk usage" {
		t.Errorf("Description() = %q, want %q", got, "Disk usage")
	}
}

func TestCommandItem_FilterValueIncludesTagsAndDescription(t *testing.T) {
	item := newCommandItem(store.Command{
		Command:     "kubectl get pods",
		Tags:        []string{"k8s"},
		Description: "List pods",
	})
	fv := item.FilterValue()
	for _, want := range []string{"kubectl get pods", "k8s", "List pods"} {
		if !strings.Contains(fv, want) {
			t.Errorf("FilterValue() = %q, missing %q", fv, want)
		}
	}
}

func TestCommandItem_Command(t *testing.T) {
	item := newCommandItem(store.Command{Command: "echo hi"})
	if got := item.Command(); got != "echo hi" {
		t.Errorf("Command() = %q, want %q", got, "echo hi")
	}
}
