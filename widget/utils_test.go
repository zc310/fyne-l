package widget

import "testing"

type testCache struct {
	values map[string]string
}

func (c *testCache) Get(key []byte) ([]byte, bool) {
	value, ok := c.values[string(key)]
	if ok {
		return []byte(value), true
	}
	return nil, false
}

func (c *testCache) Set(key, value []byte) error {
	c.values[string(key)] = string(value)
	return nil
}

func TestLastFileUsesConfiguredCache(t *testing.T) {
	c := &testCache{values: map[string]string{"file": "cached.txt"}}
	SetCache(c)
	t.Cleanup(func() { SetCache(nil) })

	lastFile := NewLastFile("file", "default.txt")
	if lastFile.Value != "cached.txt" {
		t.Fatalf("NewLastFile() = %q, want %q", lastFile.Value, "cached.txt")
	}

	lastFile.Value = "updated.txt"
	lastFile.Close()
	if got := c.values["file"]; got != "updated.txt" {
		t.Fatalf("cached value = %q, want %q", got, "updated.txt")
	}
}

func TestDefaultCacheStoresStrings(t *testing.T) {
	SetCache(nil)
	t.Cleanup(func() { SetCache(nil) })

	if err := SetCachedString("key", "value"); err != nil {
		t.Fatalf("SetCachedString() error = %v", err)
	}
	if got := GetCachedString("key", "fallback"); got != "value" {
		t.Fatalf("GetCachedString() = %q, want %q", got, "value")
	}
	if got := GetCachedString("missing", "fallback"); got != "fallback" {
		t.Fatalf("GetCachedString() missing = %q, want %q", got, "fallback")
	}
}

func TestIntEntryZeroBounds(t *testing.T) {
	nonNegative := NewNonNegativeIntEntry()
	if err := nonNegative.validate("-1"); err == nil {
		t.Fatal("NewNonNegativeIntEntry() accepted a negative value")
	}
	if err := nonNegative.validate("0"); err != nil {
		t.Fatalf("NewNonNegativeIntEntry() rejected zero: %v", err)
	}

	percentage := NewPercentageEntry()
	if err := percentage.validate("-1"); err == nil {
		t.Fatal("NewPercentageEntry() accepted a negative value")
	}
	if err := percentage.validate("0"); err != nil {
		t.Fatalf("NewPercentageEntry() rejected zero: %v", err)
	}
}

func TestNormalizeFileExtension(t *testing.T) {
	tests := map[string]string{
		"json":   ".json",
		" .txt ": ".txt",
		"":       "",
		".":      "",
	}
	for input, want := range tests {
		if got := normalizeFileExtension(input); got != want {
			t.Errorf("normalizeFileExtension(%q) = %q, want %q", input, got, want)
		}
	}
}
