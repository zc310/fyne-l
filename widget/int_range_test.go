package widget

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestIntRangeSetEntryMinWidth(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	changes := 0
	r := NewIntRange("", 123, 456, func(min, max int) { changes++ }, false)
	r.SetEntryMinWidth(96)
	for _, entry := range []*IntEntry{r.minEntry, r.maxEntry} {
		if entry.GetMinWidth() != 96 || entry.MinSize().Width < 96 {
			t.Fatal("both entries should use the configured minimum width")
		}
	}
	if min, max := r.Range(); min != 123 || max != 456 || changes != 0 {
		t.Fatal("changing width should preserve values without firing callbacks")
	}
	r.maxEntry.SetText("789")
	if min, max := r.Range(); min != 123 || max != 789 || changes != 1 {
		t.Fatal("resized entries should continue to update the range")
	}
}
