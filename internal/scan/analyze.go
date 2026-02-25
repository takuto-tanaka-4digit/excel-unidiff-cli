package scan

import (
	"sort"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type Metrics struct {
	NFCOnly            int
	NFDOnly            int
	NFCCollisions      int
	CombiningMarkPaths int
}

type Collision struct {
	NormalizedPath string
	CollidingPaths []string
}

type Analysis struct {
	Metrics    Metrics
	Collisions []Collision
}

func AnalyzePaths(paths []string) Analysis {
	analysis := Analysis{}
	groups := make(map[string]map[string]struct{})

	for _, p := range paths {
		nfc := norm.NFC.IsNormalString(p)
		nfd := norm.NFD.IsNormalString(p)
		if nfc && !nfd {
			analysis.Metrics.NFCOnly++
		}
		if nfd && !nfc {
			analysis.Metrics.NFDOnly++
		}
		if hasCombiningMark(p) {
			analysis.Metrics.CombiningMarkPaths++
		}

		nfcPath := norm.NFC.String(p)
		if _, ok := groups[nfcPath]; !ok {
			groups[nfcPath] = make(map[string]struct{})
		}
		groups[nfcPath][p] = struct{}{}
	}

	for normalized, set := range groups {
		if len(set) <= 1 {
			continue
		}
		colliding := make([]string, 0, len(set))
		for p := range set {
			colliding = append(colliding, p)
		}
		sort.Strings(colliding)
		analysis.Collisions = append(analysis.Collisions, Collision{
			NormalizedPath: normalized,
			CollidingPaths: colliding,
		})
	}
	sort.Slice(analysis.Collisions, func(i, j int) bool {
		return analysis.Collisions[i].NormalizedPath < analysis.Collisions[j].NormalizedPath
	})

	analysis.Metrics.NFCCollisions = len(analysis.Collisions)
	return analysis
}

func hasCombiningMark(s string) bool {
	for _, r := range s {
		if unicode.IsMark(r) {
			return true
		}
	}
	return false
}
