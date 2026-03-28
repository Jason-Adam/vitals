//go:build !unix

package gather

// getTermWidth is a no-op on non-Unix platforms; the caller falls back
// to the COLUMNS env var.
func getTermWidth(_ uintptr) int { return 0 }
