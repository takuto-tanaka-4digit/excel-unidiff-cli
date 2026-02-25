package policy

import (
	"path"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type CaseCollision struct {
	Normalized string
	Candidates []string
}

func NormalizeRelativePath(in string) string {
	normalizedSlash := strings.ReplaceAll(in, "\\", "/")
	cleaned := path.Clean(normalizedSlash)
	if cleaned == "." {
		cleaned = ""
	}
	cleaned = strings.TrimPrefix(cleaned, "./")
	return norm.NFC.String(cleaned)
}

func DetectCaseOnlyCollisions(paths []string) []CaseCollision {
	groups := make(map[string]map[string]struct{})
	for _, p := range paths {
		normalized := NormalizeRelativePath(p)
		lowerKey := strings.ToLower(normalized)
		if _, ok := groups[lowerKey]; !ok {
			groups[lowerKey] = make(map[string]struct{})
		}
		groups[lowerKey][normalized] = struct{}{}
	}

	keys := make([]string, 0, len(groups))
	for key, group := range groups {
		if len(group) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	collisions := make([]CaseCollision, 0, len(keys))
	for _, key := range keys {
		candidates := make([]string, 0, len(groups[key]))
		for candidate := range groups[key] {
			candidates = append(candidates, candidate)
		}
		sort.Strings(candidates)
		collisions = append(collisions, CaseCollision{
			Normalized: key,
			Candidates: candidates,
		})
	}

	return collisions
}
