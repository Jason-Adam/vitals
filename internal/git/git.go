// Package git provides git repository status information for the working directory.
package git

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/logging"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

const timeout = time.Second

// GetStatus returns the git status for the given directory, or nil if the
// directory is not a git repo or any command fails. Callers must guard against
// nil before dereferencing the result.
func GetStatus(cwd string) *model.GitStatus {
	branch, ok := getBranch(cwd)
	if !ok {
		// Not a git repo or git unavailable — fail open.
		return nil
	}

	status := &model.GitStatus{Branch: branch}

	if err := applyPorcelain(cwd, status); err != nil {
		logging.Debug("git: porcelain failed in %s: %v", cwd, err)
		return nil
	}

	// Ahead/behind: failure is non-fatal — default to 0/0 (no upstream).
	applyAheadBehind(cwd, status)

	return status
}

// getBranch runs `git rev-parse --abbrev-ref HEAD` and returns the branch name.
// Returns ("", false) when the command fails, which signals we are not in a git repo.
func getBranch(cwd string) (string, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		logging.Debug("git: rev-parse failed in %s: %v", cwd, err)
		return "", false
	}
	return strings.TrimSpace(string(out)), true
}

// applyPorcelain runs `git status --porcelain` and populates staged, modified,
// untracked, and dirty fields on the provided GitStatus.
func applyPorcelain(cwd string, status *model.GitStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 2 {
			continue
		}
		status.Dirty = true
		x := line[0] // index status
		y := line[1] // worktree status

		// Untracked: both columns are '?'
		if x == '?' && y == '?' {
			status.Untracked++
			continue
		}

		// Staged: index column is not space or '?'
		if x != ' ' && x != '?' {
			status.Staged++
		}

		// Modified in worktree: 'M' (modified) or 'D' (deleted)
		if y == 'M' || y == 'D' {
			status.Modified++
		}
	}

	return nil
}

// applyAheadBehind runs `git rev-list --left-right --count @{upstream}...HEAD`
// and sets AheadBy / BehindBy on the status. Silently ignores errors so that
// repos without an upstream configured still return a valid (zero) result.
func applyAheadBehind(cwd string, status *model.GitStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		// No upstream or other failure — leave counts at zero.
		logging.Debug("git: rev-list upstream failed in %s: %v", cwd, err)
		return
	}

	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) != 2 {
		return
	}

	behind, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	ahead, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}

	status.BehindBy = behind
	status.AheadBy = ahead
}
