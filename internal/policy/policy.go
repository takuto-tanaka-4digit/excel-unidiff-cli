package policy

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	VersionCurrent     = 1
	KeyPrecompose      = "core.precomposeunicode"
	KeyProtectHFS      = "core.protecthfs"
	DefaultPolicyFile  = ".euni.yml"
	TemplatePolicyYAML = "version: 1\ndefaults:\n  darwin:\n    core.precomposeunicode: true\n    core.protecthfs: true\n  others:\n    core.precomposeunicode: false\nsubmodules: {}\nnestedRepos: {}\n"
)

var ErrInvalidPolicy = errors.New("invalid policy")

type repoRule struct {
	CorePrecomposeUnicode *bool `yaml:"core.precomposeunicode"`
	CoreProtectHFS        *bool `yaml:"core.protecthfs"`
}

type defaultsRule struct {
	Darwin repoRule `yaml:"darwin"`
	Others repoRule `yaml:"others"`
}

type rawPolicy struct {
	Version     int                 `yaml:"version"`
	Defaults    defaultsRule        `yaml:"defaults"`
	Submodules  map[string]repoRule `yaml:"submodules"`
	NestedRepos map[string]repoRule `yaml:"nestedRepos"`
}

type Policy struct {
	Version int

	defaultsDarwin map[string]bool
	defaultsOthers map[string]bool
	submodules     map[string]map[string]bool
	nestedRepos    map[string]map[string]bool

	submoduleCollisions map[string]CaseCollision
	nestedCollisions    map[string]CaseCollision
}

func Load(path string) (Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	return Parse(b)
}

func Parse(content []byte) (Policy, error) {
	var raw rawPolicy
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(&raw); err != nil {
		return Policy{}, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}
	if raw.Version != VersionCurrent {
		return Policy{}, fmt.Errorf("%w: unsupported policy version: %d", ErrInvalidPolicy, raw.Version)
	}

	if raw.Submodules == nil {
		raw.Submodules = map[string]repoRule{}
	}
	if raw.NestedRepos == nil {
		raw.NestedRepos = map[string]repoRule{}
	}

	policy := Policy{
		Version:             raw.Version,
		defaultsDarwin:      defaultDarwin(),
		defaultsOthers:      defaultOthers(),
		submodules:          map[string]map[string]bool{},
		nestedRepos:         map[string]map[string]bool{},
		submoduleCollisions: toCollisionMap(DetectCaseOnlyCollisions(mapKeys(raw.Submodules))),
		nestedCollisions:    toCollisionMap(DetectCaseOnlyCollisions(mapKeys(raw.NestedRepos))),
	}

	if m := raw.Defaults.Darwin.toMap(); len(m) > 0 {
		policy.defaultsDarwin = merge(policy.defaultsDarwin, m)
	}
	if m := raw.Defaults.Others.toMap(); len(m) > 0 {
		policy.defaultsOthers = merge(policy.defaultsOthers, m)
	}

	for key, value := range raw.Submodules {
		normKey := NormalizeRelativePath(key)
		policy.submodules[normKey] = value.toMap()
	}
	for key, value := range raw.NestedRepos {
		normKey := NormalizeRelativePath(key)
		policy.nestedRepos[normKey] = value.toMap()
	}

	return policy, nil
}

func (r repoRule) toMap() map[string]bool {
	result := make(map[string]bool)
	if r.CorePrecomposeUnicode != nil {
		result[KeyPrecompose] = *r.CorePrecomposeUnicode
	}
	if r.CoreProtectHFS != nil {
		result[KeyProtectHFS] = *r.CoreProtectHFS
	}
	return result
}

func (p Policy) ExpectedFor(repoRelPath string, repoKind string) (expected map[string]bool, unresolvedCollision *CaseCollision) {
	normRel := NormalizeRelativePath(repoRelPath)
	lower := strings.ToLower(normRel)

	base := p.defaultsOthers
	if runtime.GOOS == "darwin" {
		base = p.defaultsDarwin
	}
	expected = clone(base)

	switch repoKind {
	case "submodule":
		if col, ok := p.submoduleCollisions[lower]; ok {
			return expected, &col
		}
		expected = merge(expected, p.submodules[normRel])
	case "nested":
		if col, ok := p.nestedCollisions[lower]; ok {
			return expected, &col
		}
		expected = merge(expected, p.nestedRepos[normRel])
	}

	return expected, nil
}

func (p Policy) SubmoduleCollisions() []CaseCollision {
	return sortedCollisions(p.submoduleCollisions)
}

func (p Policy) NestedRepoCollisions() []CaseCollision {
	return sortedCollisions(p.nestedCollisions)
}

func defaultDarwin() map[string]bool {
	return map[string]bool{
		KeyPrecompose: true,
		KeyProtectHFS: true,
	}
}

func defaultOthers() map[string]bool {
	return map[string]bool{
		KeyPrecompose: false,
	}
}

func clone(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func merge(base, override map[string]bool) map[string]bool {
	if len(override) == 0 {
		return clone(base)
	}
	out := clone(base)
	for k, v := range override {
		out[k] = v
	}
	return out
}

func toCollisionMap(in []CaseCollision) map[string]CaseCollision {
	m := make(map[string]CaseCollision, len(in))
	for _, c := range in {
		m[c.Normalized] = c
	}
	return m
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func sortedCollisions(m map[string]CaseCollision) []CaseCollision {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]CaseCollision, 0, len(keys))
	for _, key := range keys {
		out = append(out, m[key])
	}
	return out
}
