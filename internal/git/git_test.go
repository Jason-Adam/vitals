package git_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/git"
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

// TestGetStatus_NotARepo verifies GetStatus returns nil outside a git repo.
func TestGetStatus_NotARepo(t *testing.T) {
	dir := t.TempDir()
	status := git.GetStatus(dir)
	if status != nil {
		t.Fatalf("expected nil for non-git directory, got %+v", status)
	}
}

// TestGetStatus_CleanRepo verifies a clean repo with one commit reports the
// correct branch name and zero counts.
func TestGetStatus_CleanRepo(t *testing.T) {
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
// This exercises the non-fatal code path in applyAheadBehind.
func TestGetStatus_NoUpstream(t *testing.T) {
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
