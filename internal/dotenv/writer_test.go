package dotenv

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWrite_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")

	w := NewWriter(p, false)
	if err := w.Write(map[string]string{"FOO": "bar", "BAZ": "qux"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "BAZ=qux\n") {
		t.Errorf("expected BAZ=qux, got: %s", content)
	}
	if !strings.Contains(content, "FOO=bar\n") {
		t.Errorf("expected FOO=bar, got: %s", content)
	}
}

func TestWrite_FilePermissions(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")

	w := NewWriter(p, false)
	_ = w.Write(map[string]string{"KEY": "val"})

	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600, got %v", info.Mode().Perm())
	}
}

func TestWrite_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ".env")
	_ = os.WriteFile(p, []byte("OLD=data\n"), 0600)

	w := NewWriter(p, true)
	_ = w.Write(map[string]string{"NEW": "value"})

	bak, err := os.ReadFile(p + ".bak")
	if err != nil {
		t.Fatalf("backup not created: %v", err)
	}
	if string(bak) != "OLD=data\n" {
		t.Errorf("unexpected backup content: %s", bak)
	}
}

func TestQuoteValue(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"simple", "simple"},
		{"has space", `"has space"`},
		{"has\nnewline", `"has\nnewline"`},
		{`has"quote`, `"has\"quote"`},
	}
	for _, c := range cases {
		got := quoteValue(c.input)
		if got != c.want {
			t.Errorf("quoteValue(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}
