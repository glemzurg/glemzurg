package identity

import "strings"

// NormalizeSubKey converts a human-readable name into a valid SubKey identifier.
// It trims whitespace, lowercases, replaces spaces and hyphens with underscores.
func NormalizeSubKey(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}
