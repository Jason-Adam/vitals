// Package render walks config lines, calls widget functions, joins non-empty
// results with the configured separator, and writes each line to an io.Writer.
package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/render/widget"
)

// Render walks config lines, looks up widgets in the registry, joins non-empty
// results with the configured separator, and writes each line to w.
// Unknown widget names are skipped silently. Lines where all widgets return
// empty strings are skipped entirely.
// Terminal width truncation is deferred to Phase 2.
func Render(w io.Writer, ctx *model.RenderContext, cfg *config.Config) {
	sep := cfg.Style.Separator

	for _, line := range cfg.Lines {
		var parts []string
		for _, name := range line.Widgets {
			fn, ok := widget.Registry[name]
			if !ok {
				continue // skip unknown widget names silently
			}
			if s := fn(ctx, cfg); s != "" {
				parts = append(parts, s)
			}
		}
		if len(parts) == 0 {
			continue // skip lines where every widget returned empty
		}
		output := strings.Join(parts, sep)
		fmt.Fprintln(w, output)
	}
}
