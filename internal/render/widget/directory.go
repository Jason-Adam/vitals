package widget

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

var dirStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)

// Directory renders the last N path segments from ctx.Cwd, where N is
// cfg.Directory.Levels (defaulting to 1). Uses magenta bold styling.
// Returns "" when ctx.Cwd is empty.
func Directory(ctx *model.RenderContext, cfg *config.Config) string {
	if ctx.Cwd == "" {
		return ""
	}

	levels := cfg.Directory.Levels
	if levels <= 0 {
		levels = 1
	}

	segments := lastNSegments(ctx.Cwd, levels)
	return dirStyle.Render(segments)
}

// lastNSegments returns the last n path segments from a slash-delimited path,
// joined with "/". Trailing slashes are trimmed before splitting.
func lastNSegments(path string, n int) string {
	path = strings.TrimRight(path, "/")
	if path == "" {
		return ""
	}

	parts := strings.Split(path, "/")

	// Remove any empty leading segment that results from an absolute path.
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	if len(parts) == 0 {
		return ""
	}

	if n >= len(parts) {
		return strings.Join(parts, "/")
	}

	return strings.Join(parts[len(parts)-n:], "/")
}
