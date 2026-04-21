package git_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Jason-Adam/vitals/internal/git"
	"github.com/Jason-Adam/vitals/internal/model"
)

// initRepo creates a minimal git repo in dir with a configured identity so
// commits succeed in CI environments where global git config may be absent.
func initRepo(t *testing.T, dir string) {
	t.Helper()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
}

func commitFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cmd := exec.Command("git", "add", name)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "commit", "-m", "add "+name)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v\n%s", err, out)
	}
}

// isolateHome redirects HOME so cache files land in t.TempDir() and never
// pollute the real ~/.claude/plugins/vitals/.
func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

// TestGetStatus_NotARepo verifies GetStatus returns nil outside a git repo.
func TestGetStatus_NotARepo(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	status := git.GetStatus(dir)
	if status != nil {
		t.Fatalf("expected nil for non-git directory, got %+v", status)
	}
}

// TestGetStatus_CleanRepo verifies a clean repo with one commit reports the
// correct branch name and zero counts.
func TestGetStatus_CleanRepo(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status for git repo")
	}
	if status.Branch == "" {
		t.Error("expected non-empty branch name")
	}
	if status.Dirty {
		t.Error("expected clean repo to not be dirty")
	}
	if status.Modified != 0 {
		t.Errorf("expected 0 modified, got %d", status.Modified)
	}
	if status.Staged != 0 {
		t.Errorf("expected 0 staged, got %d", status.Staged)
	}
	if status.Untracked != 0 {
		t.Errorf("expected 0 untracked, got %d", status.Untracked)
	}
}

// TestGetStatus_BranchName verifies the branch name is reported correctly after
// checking out a non-default branch.
func TestGetStatus_BranchName(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "file.txt", "content")

	cmd := exec.Command("git", "checkout", "-b", "test-branch")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git checkout -b: %v\n%s", err, out)
	}

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.Branch != "test-branch" {
		t.Errorf("expected branch 'test-branch', got %q", status.Branch)
	}
}

// TestGetStatus_UntrackedFile verifies untracked files are counted and Dirty is set.
func TestGetStatus_UntrackedFile(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	// Add an untracked file — do not stage it.
	if err := os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("untracked"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Dirty {
		t.Error("expected Dirty=true with untracked file")
	}
	if status.Untracked != 1 {
		t.Errorf("expected 1 untracked, got %d", status.Untracked)
	}
	if status.Staged != 0 {
		t.Errorf("expected 0 staged, got %d", status.Staged)
	}
}

// TestGetStatus_StagedFile verifies staged files are counted correctly.
func TestGetStatus_StagedFile(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	// Create and stage a new file.
	newFile := filepath.Join(dir, "staged.txt")
	if err := os.WriteFile(newFile, []byte("staged"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	cmd := exec.Command("git", "add", "staged.txt")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Dirty {
		t.Error("expected Dirty=true with staged file")
	}
	if status.Staged != 1 {
		t.Errorf("expected 1 staged, got %d", status.Staged)
	}
}

// TestGetStatus_ModifiedFile verifies modified (unstaged) files are counted.
func TestGetStatus_ModifiedFile(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	// Modify the file without staging.
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if !status.Dirty {
		t.Error("expected Dirty=true with modified file")
	}
	if status.Modified != 1 {
		t.Errorf("expected 1 modified, got %d", status.Modified)
	}
}

// TestGetStatus_NoUpstream verifies ahead/behind are zero when no remote is configured.
func TestGetStatus_NoUpstream(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	status := git.GetStatus(dir)
	if status == nil {
		t.Fatal("expected non-nil status")
	}
	if status.AheadBy != 0 {
		t.Errorf("expected AheadBy=0 with no upstream, got %d", status.AheadBy)
	}
	if status.BehindBy != 0 {
		t.Errorf("expected BehindBy=0 with no upstream, got %d", status.BehindBy)
	}
}

// TestGetStatus_OnDiskCacheHit verifies that a second call within the TTL
// window returns the cached result without invoking git. A file is added
// between the two calls; the cached (pre-change) snapshot must be returned.
// The cache file must exist on disk after the first call.
func TestGetStatus_OnDiskCacheHit(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	first := git.GetStatus(dir)
	if first == nil {
		t.Fatal("expected non-nil status on first call")
	}

	// Verify the cache file exists on disk (one file per distinct cwd).
	pluginDir := model.PluginDir()
	matches, err := filepath.Glob(filepath.Join(pluginDir, ".git-*.json"))
	if err != nil {
		t.Fatalf("glob cache dir: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 cache file, got %d: %v", len(matches), matches)
	}

	// Add an untracked file; a cache hit must not see it.
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	second := git.GetStatus(dir)
	if second == nil {
		t.Fatal("expected non-nil status on second call")
	}
	if second.Untracked != 0 {
		t.Errorf("expected 0 untracked from cache hit, got %d", second.Untracked)
	}
	if second.Dirty {
		t.Error("expected Dirty=false from cache hit, got true")
	}
}

// TestGetStatus_OnDiskCacheExpiry verifies that a call after the TTL window
// spawns a fresh subprocess and reflects the updated state.
func TestGetStatus_OnDiskCacheExpiry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cache expiry test in short mode (requires >2s sleep)")
	}
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	first := git.GetStatus(dir)
	if first == nil {
		t.Fatal("expected non-nil status on first call")
	}

	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// Sleep past cacheTTL (2s) with a small margin.
	time.Sleep(2100 * time.Millisecond)

	second := git.GetStatus(dir)
	if second == nil {
		t.Fatal("expected non-nil status after cache expiry")
	}
	if second.Untracked != 1 {
		t.Errorf("expected 1 untracked after cache expiry, got %d", second.Untracked)
	}
}

// TestGetStatus_IndexLockPresent verifies that when .git/index.lock exists,
// GetStatus yields the subprocess and returns the cached value — even when
// that cache is stale — rather than potentially blocking a user git op.
func TestGetStatus_IndexLockPresent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping index.lock test in short mode (requires >2s sleep)")
	}
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	// Warm the cache.
	warmed := git.GetStatus(dir)
	if warmed == nil {
		t.Fatal("expected non-nil warm-up status")
	}
	if warmed.Untracked != 0 {
		t.Fatalf("warm-up expected 0 untracked, got %d", warmed.Untracked)
	}

	// Expire the cache so the guard path is exercised (stale-but-returned).
	time.Sleep(2100 * time.Millisecond)

	// Create the lock and modify the repo while the lock is held.
	lockPath := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lockPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create index.lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("new"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// With the lock present, GetStatus must return the stale cached snapshot.
	locked := git.GetStatus(dir)
	if locked == nil {
		t.Fatal("expected stale cached status when index.lock is present")
	}
	if locked.Untracked != 0 {
		t.Errorf("expected 0 untracked from cached snapshot, got %d", locked.Untracked)
	}

	// Remove the lock; the next call should refresh and see the untracked file.
	if err := os.Remove(lockPath); err != nil {
		t.Fatalf("remove lock: %v", err)
	}
	fresh := git.GetStatus(dir)
	if fresh == nil {
		t.Fatal("expected non-nil status after lock removed")
	}
	if fresh.Untracked != 1 {
		t.Errorf("expected 1 untracked after lock removed, got %d", fresh.Untracked)
	}
}

// TestGetStatus_IndexLockNoCache verifies that when the index is locked and
// no cache exists, GetStatus returns nil rather than calling git.
func TestGetStatus_IndexLockNoCache(t *testing.T) {
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	// Create the lock before the first call so no cache has been written.
	lockPath := filepath.Join(dir, ".git", "index.lock")
	if err := os.WriteFile(lockPath, []byte{}, 0o644); err != nil {
		t.Fatalf("create index.lock: %v", err)
	}

	status := git.GetStatus(dir)
	if status != nil {
		t.Errorf("expected nil when index.lock is held and no cache exists, got %+v", status)
	}
}

// TestGetStatus_UsesNoOptionalLocks deterministically verifies the
// --no-optional-locks flag is passed to git. A shim "git" binary intercepts
// the invocation, records its arguments, and forwards to the real git so the
// rest of the code path still works.
func TestGetStatus_UsesNoOptionalLocks(t *testing.T) {
	realGit, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git not in PATH")
	}
	isolateHome(t)

	shimDir := t.TempDir()
	logFile := filepath.Join(shimDir, "args.log")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$*\" >> %q\nexec %q \"$@\"\n", logFile, realGit)
	shimPath := filepath.Join(shimDir, "git")
	if err := os.WriteFile(shimPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write shim: %v", err)
	}

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", shimDir+string(os.PathListSeparator)+origPath)

	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	_ = git.GetStatus(dir)

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read args log: %v", err)
	}
	// Among many git invocations (init, config, add, commit, status), at
	// least one must carry --no-optional-locks followed by status.
	if !strings.Contains(string(data), "--no-optional-locks status --branch --porcelain=v2") {
		t.Errorf("expected --no-optional-locks in GetStatus args, got:\n%s", data)
	}
}

// TestGetStatus_NoOptionalLocksNoContention runs concurrent GetStatus calls
// alongside user git operations and asserts none of the user ops surfaces
// "Another git process seems to be running" — the exact symptom of
// contending on .git/index.lock that --no-optional-locks is meant to avoid.
func TestGetStatus_NoOptionalLocksNoContention(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping lock-contention stress test in short mode")
	}
	isolateHome(t)
	dir := t.TempDir()
	initRepo(t, dir)
	commitFile(t, dir, "README.md", "hello")

	const getStatusWorkers = 8
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	pluginDir := model.PluginDir()
	for range getStatusWorkers {
		wg.Go(func() {
			for ctx.Err() == nil {
				// Force each call to hit the subprocess path rather than cache,
				// so the test actually exercises concurrent git invocations.
				_ = os.RemoveAll(pluginDir)
				_ = git.GetStatus(dir)
			}
		})
	}

	// classify runs a git op and reports: nil success, contention (the error
	// this test is asserting against), or any other error (fails the test so
	// environmental issues don't silently make the stress test pass).
	classify := func(label, out string, err error) (contended bool) {
		if err == nil {
			return false
		}
		if strings.Contains(out, "Another git process") {
			return true
		}
		t.Errorf("unexpected %s failure: %v\n%s", label, err, out)
		return false
	}

	var contention atomic.Int32
	const userOps = 200
	for i := range userOps {
		fname := fmt.Sprintf("f%d", i)
		if err := os.WriteFile(filepath.Join(dir, fname), []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		addCmd := exec.Command("git", "add", fname)
		addCmd.Dir = dir
		addOut, addErr := addCmd.CombinedOutput()
		if classify("git add", string(addOut), addErr) {
			contention.Add(1)
		}
		resetCmd := exec.Command("git", "reset")
		resetCmd.Dir = dir
		resetOut, resetErr := resetCmd.CombinedOutput()
		if classify("git reset", string(resetOut), resetErr) {
			contention.Add(1)
		}
	}

	cancel()
	wg.Wait()

	if c := contention.Load(); c > 0 {
		t.Errorf("observed %d lock contention errors from user git ops", c)
	}
}
