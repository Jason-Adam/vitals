package widget

import (
	"fmt"
	"strings"

	"github.com/kylesnowschwartz/tail-claude-hud/internal/config"
	"github.com/kylesnowschwartz/tail-claude-hud/internal/model"
)

// Tokens renders a compact breakdown of token usage from stdin CurrentUsage.
// Format: "45.1k in · 12.3k cache" (cache = cacheCreation + cacheRead combined).
//
// Output tokens are not included: the stdin JSON current_usage field only provides
// input_tokens, cache_creation_input_tokens, and cache_read_input_tokens. There is
// no output_tokens field available at the time this widget runs.
//
// Returns "" when all token counts are zero.
func Tokens(ctx *model.RenderContext, cfg *config.Config) string {
	in := ctx.InputTokens
	cacheCreate := ctx.CacheCreation
	cacheRead := ctx.CacheRead

	if in == 0 && cacheCreate == 0 && cacheRead == 0 {
		return ""
	}

	// Combine cache creation and cache read into a single "cache" figure for brevity.
	cache := cacheCreate + cacheRead

	var parts []string
	parts = append(parts, fmt.Sprintf("%s in", formatTokenCount(in)))
	if cache > 0 {
		parts = append(parts, fmt.Sprintf("%s cache", formatTokenCount(cache)))
	}

	return dimStyle.Render(strings.Join(parts, " · "))
}
