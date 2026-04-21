package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Jason-Adam/vitals/internal/git"
	"github.com/Jason-Adam/vitals/internal/model"
)

// BenchmarkGetStatus_WarmCache measures the on-disk cache-hit path — the
// overwhelmingly common case during active sessions. No subprocess is spawned;
// cost is a single file read + JSON unmarshal.
//
// TARGET: single-digit microseconds on Apple M-series (darwin/arm64).
func BenchmarkGetStatus_WarmCache(b *testing.B) {
	b.ReportAllocs()
	b.Setenv("HOME", b.TempDir())

	dir := b.TempDir()
	initGitRepo(b, dir)
	// Prime the cache so every iteration is a hit.
	_ = git.GetStatus(dir)

	for b.Loop() {
		_ = git.GetStatus(dir)
	}
}

// BenchmarkGetStatus_ColdCache measures the cache-miss path: each iteration
// removes the cache file, forcing a fresh git subprocess. Reflects the worst
// case when the TTL has expired or the process is new.
//
// TARGET: under 10ms per operation on Apple M-series.
func BenchmarkGetStatus_ColdCache(b *testing.B) {
	b.ReportAllocs()
	b.Setenv("HOME", b.TempDir())

	dir := b.TempDir()
	initGitRepo(b, dir)

	pluginDir := model.PluginDir()
	for b.Loop() {
		b.StopTimer()
		_ = os.RemoveAll(pluginDir)
		b.StartTimer()
		_ = git.GetStatus(dir)
	}
}

// BenchmarkGetStatus_ColdCache_WithChanges benchmarks the cache-miss path
// against a repo with staged, modified, and untracked files — a realistic
// working-session state with non-empty porcelain output to parse.
func BenchmarkGetStatus_ColdCache_WithChanges(b *testing.B) {
	b.ReportAllocs()
	b.Setenv("HOME", b.TempDir())

	dir := b.TempDir()
	initGitRepo(b, dir)

	writeFile(b, filepath.Join(dir, "staged.txt"), "staged content\n")
	gitCmd(b, dir, "add", "staged.txt")
	writeFile(b, filepath.Join(dir, "untracked.txt"), "untracked\n")
	writeFile(b, filepath.Join(dir, "hello.txt"), "modified content\n")

	pluginDir := model.PluginDir()
	for b.Loop() {
		b.StopTimer()
		_ = os.RemoveAll(pluginDir)
		b.StartTimer()
		_ = git.GetStatus(dir)
	}
}

// BenchmarkGetStatus_NonGitDir measures the fast-fail path: GetStatus called
// on a directory that is not a git repository. Non-repo results are not
// cached (per design — the "do not clobber" rule) so every call pays the
// subprocess cost.
func BenchmarkGetStatus_NonGitDir(b *testing.B) {
	b.ReportAllocs()
	b.Setenv("HOME", b.TempDir())

	dir := b.TempDir()

	for b.Loop() {
		_ = git.GetStatus(dir)
	}
}

// initGitRepo initialises a minimal git repository at dir with a single commit.
func initGitRepo(b *testing.B, dir string) {
	b.Helper()

	gitCmd(b, dir, "init", "-b", "main")
	gitCmd(b, dir, "config", "user.email", "bench@example.com")
	gitCmd(b, dir, "config", "user.name", "Benchmarker")

	writeFile(b, filepath.Join(dir, "hello.txt"), "hello\n")
	gitCmd(b, dir, "add", "hello.txt")
	gitCmd(b, dir, "commit", "-m", "init")
}

// gitCmd runs a git command in dir, failing the benchmark on error.
func gitCmd(b *testing.B, dir string, args ...string) {
	b.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		b.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// writeFile writes content to path, failing the benchmark on error.
func writeFile(b *testing.B, path, content string) {
	b.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		b.Fatalf("writeFile %s: %v", path, err)
	}
}
