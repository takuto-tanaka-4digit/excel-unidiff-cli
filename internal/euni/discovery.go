package euni

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/policy"
	"github.com/takuto-tanaka-4digit/excel-unidiff-cli/internal/report"
)

type repoTarget struct {
	AbsPath string
	RelPath string
	Kind    string
}

type discoveryResult struct {
	RootCanonical string
	Repos         []repoTarget
	Findings      []report.Result
	Errors        []report.ErrorItem
}

func discoverRepositories(repoArg string, recursive bool) discoveryResult {
	rootPath, rootErr := canonicalPath(repoArg)
	if rootErr != nil {
		return discoveryResult{Errors: []report.ErrorItem{errorItem("UG001", "repository path is not accessible", strPtr(repoArg), nil)}}
	}

	rootTop, rootGitDir, err := resolveRepoPaths(repoArg)
	if err != nil {
		return discoveryResult{Errors: []report.ErrorItem{errorItem("UG001", "--repo is not a valid Git repository", strPtr(repoArg), nil)}}
	}

	if !isWithinBoundary(rootPath, rootTop) || !isWithinBoundary(rootPath, rootGitDir) {
		return discoveryResult{Errors: []report.ErrorItem{errorItem("UG010", "gitdir/top-level is outside --repo boundary", strPtr(rootTop), nil)}}
	}

	result := discoveryResult{
		RootCanonical: rootTop,
		Repos: []repoTarget{{
			AbsPath: rootTop,
			RelPath: ".",
			Kind:    "root",
		}},
	}

	if !recursive {
		return result
	}

	seen := map[string]struct{}{rootTop: {}}

	submodules, subErr := gitSubmoduleStatus(rootTop)
	if subErr != nil {
		result.Errors = append(result.Errors, errorItem("UG002", "failed to inspect submodules", strPtr(rootTop), nil))
		return result
	}
	for _, sm := range submodules {
		rel := policy.NormalizeRelativePath(sm.Path)
		if sm.Uninitialized {
			result.Errors = append(result.Errors, errorItem("UG006", "submodule is not initialized or inaccessible", strPtr(rootTop), strPtr(rel)))
			continue
		}
		candidate := filepath.Join(rootTop, sm.Path)
		if errItem := addRepoCandidate(&result, seen, rootTop, candidate, rel, "submodule"); errItem != nil {
			result.Errors = append(result.Errors, *errItem)
		}
	}

	nestedFindings, nestedErrors := discoverNestedRepoCandidates(rootTop, seen, &result.Repos)
	result.Findings = append(result.Findings, nestedFindings...)
	result.Errors = append(result.Errors, nestedErrors...)
	return result
}

func addRepoCandidate(result *discoveryResult, seen map[string]struct{}, rootCanonical, candidate, relPath, kind string) *report.ErrorItem {
	top, gitDir, err := resolveRepoPaths(candidate)
	if err != nil {
		e := errorItem("UG002", "failed to resolve nested git repository", strPtr(candidate), nil)
		return &e
	}

	if !isWithinBoundary(rootCanonical, top) || !isWithinBoundary(rootCanonical, gitDir) {
		e := errorItem("UG010", "gitdir/top-level is outside --repo boundary", strPtr(top), nil)
		return &e
	}

	if _, ok := seen[top]; ok {
		return nil
	}
	seen[top] = struct{}{}

	if relPath == "" {
		rel, err := filepath.Rel(rootCanonical, top)
		if err != nil {
			relPath = policy.NormalizeRelativePath(top)
		} else {
			relPath = policy.NormalizeRelativePath(rel)
		}
	}

	result.Repos = append(result.Repos, repoTarget{
		AbsPath: top,
		RelPath: relPath,
		Kind:    kind,
	})
	return nil
}

func discoverNestedRepoCandidates(rootCanonical string, seen map[string]struct{}, repos *[]repoTarget) ([]report.Result, []report.ErrorItem) {
	var findings []report.Result
	var errors []report.ErrorItem

	walkErr := filepath.WalkDir(rootCanonical, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			errors = append(errors, errorItem("UG002", "failed during repository walk", strPtr(path), nil))
			return nil
		}

		rel, _ := filepath.Rel(rootCanonical, path)
		relNorm := policy.NormalizeRelativePath(rel)
		if relNorm == ".git" || strings.HasPrefix(relNorm, ".git/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(relNorm, ".git/modules/") || strings.HasPrefix(relNorm, ".git/worktrees/") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if isSymlink(d) {
			p := policy.NormalizeRelativePath(rel)
			findings = append(findings, resultItem("UG012", "environment", "symbolic link/reparse entry detected", strPtr(p), "path", nil, nil, nil, nil))
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if path != rootCanonical {
			gitMarker := filepath.Join(path, ".git")
			if _, statErr := os.Lstat(gitMarker); statErr == nil {
				repoTop, gitDir, resolveErr := resolveRepoPaths(path)
				if resolveErr != nil {
					errors = append(errors, errorItem("UG002", "failed to resolve nested git repository", strPtr(path), nil))
					return nil
				}
				if !isWithinBoundary(rootCanonical, repoTop) || !isWithinBoundary(rootCanonical, gitDir) {
					errors = append(errors, errorItem("UG010", "gitdir/top-level is outside --repo boundary", strPtr(repoTop), nil))
					return nil
				}
				if _, ok := seen[repoTop]; !ok {
					seen[repoTop] = struct{}{}
					relRepo, relErr := filepath.Rel(rootCanonical, repoTop)
					if relErr != nil {
						relRepo = repoTop
					}
					*repos = append(*repos, repoTarget{
						AbsPath: repoTop,
						RelPath: policy.NormalizeRelativePath(relRepo),
						Kind:    "nested",
					})
				}
			}
		}

		return nil
	})
	if walkErr != nil {
		errors = append(errors, errorItem("UG002", "failed to complete nested repository walk", strPtr(rootCanonical), nil))
	}

	return findings, errors
}

func resolveRepoPaths(path string) (topCanonical string, gitDirCanonical string, err error) {
	top, err := gitShowTopLevel(path)
	if err != nil {
		return "", "", err
	}
	gitDir, err := gitAbsoluteGitDir(path)
	if err != nil {
		return "", "", err
	}
	topCanonical, err = canonicalPath(top)
	if err != nil {
		return "", "", fmt.Errorf("canonical top-level failed: %w", err)
	}
	gitDirCanonical, err = canonicalPath(gitDir)
	if err != nil {
		return "", "", fmt.Errorf("canonical gitdir failed: %w", err)
	}
	return topCanonical, gitDirCanonical, nil
}

func canonicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		// Existing repositories should resolve; fallback keeps behavior for non-existing paths.
		real = abs
	}
	real = filepath.Clean(real)
	normalized := filepath.ToSlash(real)
	if runtime.GOOS == "windows" {
		normalized = normalizeWindowsCanonical(normalized)
	}
	return normalized, nil
}

func isWithinBoundary(rootCanonical, candidateCanonical string) bool {
	rootCanon := rootCanonical
	candidateCanon := candidateCanonical
	if runtime.GOOS == "windows" {
		rootCanon = strings.ToLower(normalizeWindowsCanonical(rootCanon))
		candidateCanon = strings.ToLower(normalizeWindowsCanonical(candidateCanon))
	}
	root := filepath.FromSlash(rootCanon)
	candidate := filepath.FromSlash(candidateCanon)
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if rel == ".." {
		return false
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	if filepath.IsAbs(rel) {
		return false
	}
	return true
}

func isSymlink(d fs.DirEntry) bool {
	return d.Type()&os.ModeSymlink != 0
}

func normalizeWindowsCanonical(path string) string {
	v := strings.ReplaceAll(path, "\\\\", "/")
	if strings.HasPrefix(v, "//?/UNC/") {
		v = "//" + strings.TrimPrefix(v, "//?/UNC/")
	} else if strings.HasPrefix(v, "//?/") {
		v = strings.TrimPrefix(v, "//?/")
	}

	if strings.HasPrefix(v, "//") {
		rest := strings.TrimPrefix(v, "//")
		rest = strings.ReplaceAll(rest, "//", "/")
		v = "//" + rest
	} else {
		v = strings.ReplaceAll(v, "//", "/")
	}

	if len(v) >= 2 && v[1] == ':' {
		v = strings.ToUpper(v[:1]) + v[1:]
	}
	return v
}
