// Package git provides git repository status information for the working directory.
package git

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Jason-Adam/vitals/internal/logging"
	"github.com/Jason-Adam/vitals/internal/model"
)

const (
	// timeout is a hard ceiling on the git subprocess. The statusline budget is
	// single-digit ms on healthy repos; 200ms protects against pathological
	// fsmonitor/NFS cases without dominating a tick.
	timeout = 200 * time.Millisecond

	// cacheTTL is how long a fetched status is considered fresh. 2s covers
	// typing bursts while keeping the worst-case staleness small enough that
	// post-commit state (ahead/behind, dirty->clean) feels live.
	cacheTTL = 2 * time.Second

	// cacheFileTTL is how long an untouched cache file must age before the
	// sweep deletes it. 30 days covers any realistic gap between sessions.
	cacheFileTTL = 30 * 24 * time.Hour

	// sweepOdds controls how often a successful save triggers a stale-file
	// sweep. At 1-in-100, sweeps run often enough to keep the directory bounded
	// but rarely enough that the extra readdir cost is amortised away.
	sweepOdds = 100
)

// cacheEntry is the JSON structure persisted per cwd.
type cacheEntry struct {
	CWD       string           `json:"cwd"`
	Timestamp time.Time        `json:"timestamp"`
	Status    *model.GitStatus `json:"status"`
}

// GetStatus returns the git status for cwd, or nil if cwd is not a git repo
// or all fetch/cache paths fail. Callers must guard against nil.
//
// The function is structured to never block the statusline on a stuck or
// contended git: it prefers a fresh on-disk cache, yields to any user git op
// that is holding .git/index.lock, and bounds the subprocess at 200ms. On
// any fetch failure, an existing stale cache is returned in preference to nil.
func GetStatus(cwd string) *model.GitStatus {
	cached, haveCache := loadCache(cwd)
	if haveCache && time.Since(cached.Timestamp) < cacheTTL {
		return cached.Status
	}

	// Yield to concurrent user git operations. If .git/index.lock is present,
	// our subprocess could block the user and we prefer stale data. This also
	// avoids competing with commit/add/checkout for the lock. Worktrees point
	// .git at a file whose `gitdir:` line names the real gitdir under the
	// parent repo, which is where the worktree-specific index.lock lives.
	if gitDir := findGitDir(cwd); gitDir != "" {
		if _, err := os.Stat(filepath.Join(gitDir, "index.lock")); err == nil {
			logging.Debug("git: index.lock present at %s — yielding subprocess", gitDir)
			if haveCache {
				return cached.Status
			}
			return nil
		}
	}

	status, ok := fetchStatus(cwd)
	if !ok {
		// Any failure (timeout, exec error, non-repo): return stale cache if
		// present rather than clobbering a previously-good snapshot.
		if haveCache {
			return cached.Status
		}
		return nil
	}

	_ = saveCache(cwd, status)
	return status
}

// fetchStatus runs a single git subprocess and parses the porcelain v2 output.
// Returns (status, true) on success, (nil, false) on any error (including
// timeout and non-repo exit).
//
// --no-optional-locks prevents git status from taking .git/index.lock to
// refresh the stat cache. Without it, a vitals tick running in a tight loop
// can block a concurrent `git commit`/`add`/`checkout` with "Another git
// process seems to be running in this repository."
func fetchStatus(cwd string) (*model.GitStatus, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "--no-optional-locks", "status", "--branch", "--porcelain=v2")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			logging.Debug("git: porcelain=v2 timed out after %s in %s", timeout, cwd)
		} else {
			logging.Debug("git: porcelain=v2 failed in %s: %v", cwd, err)
		}
		return nil, false
	}
	return parsePorcelainV2(string(out)), true
}

// findGitDir walks upward from cwd looking for .git (either a directory, the
// common case, or a regular file, used by worktrees). When .git is a file it
// contains a `gitdir: <path>` pointer to the real git directory — typically
// <main-repo>/.git/worktrees/<name>/ — and that is where the worktree-specific
// index.lock lives. Returns "" when no .git is found.
func findGitDir(cwd string) string {
	dir := cwd
	for {
		candidate := filepath.Join(dir, ".git")
		info, err := os.Stat(candidate)
		if err == nil {
			if info.IsDir() {
				return candidate
			}
			data, err := os.ReadFile(candidate)
			if err != nil {
				return ""
			}
			line := strings.TrimSpace(string(data))
			const prefix = "gitdir: "
			if !strings.HasPrefix(line, prefix) {
				return ""
			}
			gitDir := strings.TrimPrefix(line, prefix)
			if !filepath.IsAbs(gitDir) {
				gitDir = filepath.Join(dir, gitDir)
			}
			return gitDir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// cacheFilePath returns the on-disk cache file for cwd. The filename is a
// short SHA-256 prefix of the cwd; a 48-bit prefix makes same-machine
// collisions vanishingly unlikely, and loadCache verifies the embedded CWD
// field as a last-resort collision guard.
func cacheFilePath(cwd string) string {
	sum := sha256.Sum256([]byte(cwd))
	name := fmt.Sprintf(".git-%x.json", sum[:6])
	return filepath.Join(model.PluginDir(), name)
}

// loadCache reads and parses the cache file for cwd. Returns (entry, true)
// when a valid cache is present, (nil, false) otherwise. A cache with a
// mismatched CWD field is treated as absent (hash collision guard).
func loadCache(cwd string) (*cacheEntry, bool) {
	data, err := os.ReadFile(cacheFilePath(cwd))
	if err != nil {
		return nil, false
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}
	if entry.CWD != cwd {
		return nil, false
	}
	return &entry, true
}

// saveCache writes the cache file atomically (temp file + rename) so concurrent
// readers never see a partial write. Best-effort: errors are returned but
// callers ignore them — a failed write at worst costs the next caller a
// subprocess.
func saveCache(cwd string, status *model.GitStatus) error {
	dir := model.PluginDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entry := cacheEntry{
		CWD:       cwd,
		Timestamp: time.Now(),
		Status:    status,
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	target := cacheFilePath(cwd)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		return err
	}
	// math/rand/v2's top-level functions use a per-process random source that
	// is seeded from the OS on startup, so the 1-in-N sampling holds across
	// the one-shot-per-tick process model (unlike v1 pre-Go 1.20).
	if rand.IntN(sweepOdds) == 0 {
		sweepStaleCacheFiles(dir)
	}
	return nil
}

// sweepStaleCacheFiles removes cache files older than cacheFileTTL. Best-effort:
// any error aborts the sweep without surfacing. Same pattern as
// transcript.sweepStaleStateFiles.
func sweepStaleCacheFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-cacheFileTTL)
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, ".git-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(dir, name)) //nolint:errcheck
		}
	}
}

// parsePorcelainV2 parses the output of `git status --branch --porcelain=v2`
// and returns a populated GitStatus. The format is documented in git-status(1).
//
// Header lines start with "# " and carry branch/upstream metadata:
//
//	# branch.head <name>
//	# branch.ab +<ahead> -<behind>
//
// Entry lines describe file changes:
//
//	1 <XY> ... (ordinary changed entry)
//	2 <XY> ... (renamed/copied entry)
//	? <path>  (untracked file)
//	u <XY> ... (unmerged entry)
func parsePorcelainV2(output string) *model.GitStatus {
	status := &model.GitStatus{}

	for _, line := range strings.Split(output, "\n") {
		if len(line) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(line, "# branch.head "):
			// Branch name; "(detached)" when HEAD is detached.
			status.Branch = strings.TrimPrefix(line, "# branch.head ")

		case strings.HasPrefix(line, "# branch.ab "):
			// Ahead/behind counts relative to upstream: "+N -M"
			parts := strings.Fields(strings.TrimPrefix(line, "# branch.ab "))
			if len(parts) == 2 {
				if n, err := strconv.Atoi(strings.TrimPrefix(parts[0], "+")); err == nil {
					status.AheadBy = n
				}
				if n, err := strconv.Atoi(strings.TrimPrefix(parts[1], "-")); err == nil {
					status.BehindBy = n
				}
			}

		case line[0] == '?':
			// Untracked file.
			status.Untracked++
			status.Dirty = true

		case line[0] == '1' || line[0] == '2' || line[0] == 'u':
			// Ordinary change, rename/copy, or unmerged entry.
			// Field layout: <type> <XY> ...
			// XY is at position 2-3: X = index (staged) status, Y = worktree status.
			if len(line) < 4 {
				continue
			}
			status.Dirty = true
			x := line[2] // index column
			y := line[3] // worktree column

			// Staged: index column is not '.' (unchanged) or '?'
			if x != '.' && x != '?' {
				status.Staged++
			}

			// Modified in worktree: 'M' (modified) or 'D' (deleted)
			if y == 'M' || y == 'D' {
				status.Modified++
			}
		}
	}

	return status
}
