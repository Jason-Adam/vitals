package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/preset"
)

// writeTempTranscript creates a temporary .jsonl transcript file and returns its path.
func writeTempTranscript(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "transcript-*.jsonl")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestReadFromFile_EnvFallback(t *testing.T) {
	path := writeTempTranscript(t, `{"type":"init"}`)
	t.Setenv("CLAUDE_TRANSCRIPT_PATH", path)

	data, err := readFromFile()
	if err != nil {
		t.Fatalf("readFromFile: %v", err)
	}
	if data == nil {
		t.Fatal("expected non-nil StdinData")
	}
	if data.TranscriptPath != path {
		t.Errorf("TranscriptPath: got %q, want %q", data.TranscriptPath, path)
	}
	if data.Cwd == "" {
		t.Error("expected Cwd to be populated")
	}
}

func TestReadFromFile_MissingFile_ReturnsError(t *testing.T) {
	t.Setenv("CLAUDE_TRANSCRIPT_PATH", "/nonexistent/path/transcript.jsonl")

	_, err := readFromFile()
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

// TestEncodePath verifies Claude Code's path encoding: /, ., and _ all become -.
func TestEncodePath(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{
			input: "/Users/kyle/Code/proj",
			want:  "-Users-kyle-Code-proj",
		},
		{
			input: "/Users/kyle/.config",
			want:  "-Users-kyle--config",
		},
		{
			input: "/Users/kyle/my_project",
			want:  "-Users-kyle-my-project",
		},
		{
			input: "/Users/kyle/.claude/worktrees/agent_abc123",
			want:  "-Users-kyle--claude-worktrees-agent-abc123",
		},
		{
			input: "/private/tmp/some_dir",
			want:  "-private-tmp-some-dir",
		},
	}

	for _, tc := range cases {
		got := encodePath(tc.input)
		if got != tc.want {
			t.Errorf("encodePath(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestFindCurrentTranscript_NoFiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}
	projectDir := filepath.Join(tmp, ".claude", "projects", encodePath(cwd))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err = findCurrentTranscript()
	if err == nil {
		t.Error("expected error when no .jsonl files present, got nil")
	}
}

func TestResolvePreset_BuiltinName(t *testing.T) {
	// "default" is always a built-in preset.
	names := preset.BuiltinNames()
	if len(names) == 0 {
		t.Skip("no built-in presets defined")
	}
	p, err := resolvePreset(names[0])
	if err != nil {
		t.Fatalf("resolvePreset(%q): %v", names[0], err)
	}
	if p.Name == "" {
		t.Error("expected non-empty preset Name")
	}
}

func TestResolvePreset_UnknownName(t *testing.T) {
	_, err := resolvePreset("this-preset-definitely-does-not-exist")
	if err == nil {
		t.Fatal("expected error for unknown preset, got nil")
	}
	if !strings.Contains(err.Error(), "unknown preset") {
		t.Errorf("error message should mention 'unknown preset', got: %v", err)
	}
}

func TestResolvePreset_FilePath(t *testing.T) {
	dir := t.TempDir()
	presetFile := filepath.Join(dir, "my-preset.toml")
	content := `
[[line]]
widgets = ["model", "context"]

[style]
separator = " | "
`
	if err := os.WriteFile(presetFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write preset file: %v", err)
	}

	p, err := resolvePreset(presetFile)
	if err != nil {
		t.Fatalf("resolvePreset(%q): %v", presetFile, err)
	}
	if p.Name != "my-preset" {
		t.Errorf("Name: got %q, want %q", p.Name, "my-preset")
	}
	if p.Separator != " | " {
		t.Errorf("Separator: got %q, want %q", p.Separator, " | ")
	}
}

func TestResolvePreset_TomlSuffix(t *testing.T) {
	dir := t.TempDir()
	// A path ending in .toml without "/" should still be treated as a file path.
	presetFile := filepath.Join(dir, "custom.toml")
	if err := os.WriteFile(presetFile, []byte(`[[line]]\nwidgets = ["model"]`), 0o644); err != nil {
		t.Fatalf("write preset file: %v", err)
	}

	// Verify the resolution logic routes to LoadFromFile for .toml paths.
	// (The file content is minimal so it may not parse perfectly, but errors
	//  should come from TOML parsing, not "unknown preset".)
	_, err := resolvePreset(presetFile)
	if err != nil && strings.Contains(err.Error(), "unknown preset") {
		t.Error("a .toml-suffixed path should route to LoadFromFile, not name lookup")
	}
}

func TestFindCurrentTranscript_ReturnsNewest(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil {
		cwd = resolved
	}
	projectDir := filepath.Join(tmp, ".claude", "projects", encodePath(cwd))
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	olderPath := filepath.Join(projectDir, "older.jsonl")
	newerPath := filepath.Join(projectDir, "newer.jsonl")
	for _, p := range []string{olderPath, newerPath} {
		if err := os.WriteFile(p, []byte("{}"), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}

	// Make newerPath demonstrably newer.
	newerInfo, _ := os.Stat(newerPath)
	newTime := newerInfo.ModTime().Add(2e9) // +2s
	if err := os.Chtimes(newerPath, newTime, newTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	got, err := findCurrentTranscript()
	if err != nil {
		t.Fatalf("findCurrentTranscript: %v", err)
	}
	if got != newerPath {
		t.Errorf("got %q, want %q", got, newerPath)
	}
}
