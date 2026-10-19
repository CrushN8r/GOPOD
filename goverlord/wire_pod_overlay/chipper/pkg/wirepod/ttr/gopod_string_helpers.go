package wirepod_ttr

import "strings"

func gopodNonEmptyStrings(items ...string) []string {
	var out []string
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			out = append(out, item)
		}
	}
	return out
}
