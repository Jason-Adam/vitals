// Package eval provides tools for inspecting rendered ANSI output from the
// statusline. Parse() decodes escape sequences into structured segments so
// downstream code can reason about colors and attributes without dealing with
// raw escape bytes.
package eval

import "strings"

// ColorType identifies which color space a Color value lives in.
type ColorType int

const (
	// ColorDefault means no explicit color was set (terminal default).
	ColorDefault ColorType = iota
	// ColorANSI16 is one of the 16 standard ANSI colors (indices 0-15).
	ColorANSI16
	// ColorXterm256 is one of the 256 xterm palette colors (indices 0-255).
	ColorXterm256
	// ColorTruecolor is a 24-bit RGB color.
	ColorTruecolor
)

// Color holds a resolved terminal color value.
type Color struct {
	Type  ColorType
	Index int   // valid for ColorANSI16 and ColorXterm256
	R     uint8 // valid for ColorTruecolor
	G     uint8 // valid for ColorTruecolor
	B     uint8 // valid for ColorTruecolor
}

// StyledSegment is a run of visible text with uniform styling.
type StyledSegment struct {
	Text  string
	Fg    Color
	Bg    Color
	Bold  bool
	Faint bool
}

// style tracks the current SGR state while walking the input.
type style struct {
	fg    Color
	bg    Color
	bold  bool
	faint bool
}

func (s style) equal(o style) bool {
	return s.fg == o.fg && s.bg == o.bg && s.bold == o.bold && s.faint == o.faint
}

// Parse walks input, extracts SGR escape sequences, and returns a slice of
// StyledSegment values. Adjacent segments with identical styling are merged;
// segments with empty text are dropped.
func Parse(input string) []StyledSegment {
	var segments []StyledSegment
	var cur style
	var textBuf strings.Builder

	// flush emits the accumulated text as a segment then resets the builder.
	flush := func() {
		t := textBuf.String()
		textBuf.Reset()
		if t == "" {
			return
		}
		seg := StyledSegment{
			Text:  t,
			Fg:    cur.fg,
			Bg:    cur.bg,
			Bold:  cur.bold,
			Faint: cur.faint,
		}
		// Merge with previous segment when styling is identical.
		if len(segments) > 0 {
			prev := &segments[len(segments)-1]
			if prev.Fg == seg.Fg && prev.Bg == seg.Bg && prev.Bold == seg.Bold && prev.Faint == seg.Faint {
				prev.Text += seg.Text
				return
			}
		}
		segments = append(segments, seg)
	}

	i := 0
	for i < len(input) {
		// ESC byte.
		if input[i] != 0x1b {
			textBuf.WriteByte(input[i])
			i++
			continue
		}

		// Need at least ESC + '['.
		if i+1 >= len(input) || input[i+1] != '[' {
			// Not a CSI sequence — treat as literal.
			textBuf.WriteByte(input[i])
			i++
			continue
		}

		// Find the terminating byte of the CSI sequence.
		j := i + 2
		for j < len(input) && (input[j] < 0x40 || input[j] > 0x7e) {
			j++
		}
		if j >= len(input) {
			// Unterminated sequence — consume what we have.
			i = j
			break
		}

		terminator := input[j]
		paramStr := input[i+2 : j]
		i = j + 1

		// Only process SGR sequences (terminator == 'm').
		if terminator != 'm' {
			continue
		}

		// Parse semicolon-separated parameters.
		params := parseSGRParams(paramStr)

		// Flush accumulated text under the current style before applying the new
		// escape sequence. Order matters: flush emits with the pre-sequence style.
		flush()
		applyParams(params, &cur)
	}

	flush()
	return segments
}

// parseSGRParams splits a raw SGR parameter string into integer slices.
// An empty string yields a single zero (implicit reset).
func parseSGRParams(s string) []int {
	if s == "" {
		return []int{0}
	}
	parts := strings.Split(s, ";")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		out = append(out, atoi(p))
	}
	return out
}

// applyParams applies a slice of SGR parameters to the current style.
// Parameters are consumed left-to-right; multi-byte sequences like 38;5;N
// advance the index by more than one.
func applyParams(params []int, cur *style) {
	for i := 0; i < len(params); i++ {
		p := params[i]
		switch {
		case p == 0:
			cur.fg = Color{}
			cur.bg = Color{}
			cur.bold = false
			cur.faint = false
		case p == 1:
			cur.bold = true
		case p == 2:
			cur.faint = true
		case p == 22:
			cur.bold = false
			cur.faint = false
		case p == 39:
			cur.fg = Color{}
		case p == 49:
			cur.bg = Color{}
		case p >= 30 && p <= 37:
			cur.fg = Color{Type: ColorANSI16, Index: p - 30}
		case p >= 40 && p <= 47:
			cur.bg = Color{Type: ColorANSI16, Index: p - 40}
		case p >= 90 && p <= 97:
			cur.fg = Color{Type: ColorANSI16, Index: p - 90 + 8}
		case p >= 100 && p <= 107:
			cur.bg = Color{Type: ColorANSI16, Index: p - 100 + 8}
		case p == 38:
			if i+1 < len(params) {
				switch params[i+1] {
				case 5:
					if i+2 < len(params) {
						cur.fg = Color{Type: ColorXterm256, Index: params[i+2]}
						i += 2
					}
				case 2:
					if i+4 < len(params) {
						cur.fg = Color{
							Type: ColorTruecolor,
							R:    uint8(params[i+2]),
							G:    uint8(params[i+3]),
							B:    uint8(params[i+4]),
						}
						i += 4
					}
				}
			}
		case p == 48:
			if i+1 < len(params) {
				switch params[i+1] {
				case 5:
					if i+2 < len(params) {
						cur.bg = Color{Type: ColorXterm256, Index: params[i+2]}
						i += 2
					}
				case 2:
					if i+4 < len(params) {
						cur.bg = Color{
							Type: ColorTruecolor,
							R:    uint8(params[i+2]),
							G:    uint8(params[i+3]),
							B:    uint8(params[i+4]),
						}
						i += 4
					}
				}
			}
		}
	}
}

// atoi converts an ASCII decimal string to int, returning 0 on any error.
func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}
