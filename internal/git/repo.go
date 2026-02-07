package git

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// RepoDirName derives a filesystem-safe name from a repo URL.
// "https://github.com/anthropics/skills" -> "anthropics-skills"
func RepoDirName(repoURL string) string {
	// Try to parse as URL first
	u, err := url.Parse(repoURL)
	if err == nil && u.Host != "" {
		// Use last two path segments: org/repo
		p := strings.TrimSuffix(u.Path, ".git")
		p = strings.Trim(p, "/")
		parts := strings.Split(p, "/")
		if len(parts) >= 2 {
			return sanitizeDirName(parts[len(parts)-2] + "-" + parts[len(parts)-1])
		}
		if len(parts) == 1 {
			return sanitizeDirName(parts[0])
		}
	}
	// Fallback: use the whole string, sanitized
	return sanitizeDirName(repoURL)
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func sanitizeDirName(s string) string {
	return unsafeChars.ReplaceAllString(s, "-")
}

// RepoInstallOptions for installing a skill via repo sparse checkout.
type RepoInstallOptions struct {
	RepoURL   string // git clone URL
	Path      string // subdirectory in repo (optional, "" = repo root)
	RepoDir   string // full path to repo clone (e.g., ~/.lazyas/repos/anthropics-skills)
	SkillName string // skill name
	SkillLink string // full path to symlink target (e.g., ~/.lazyas/skills/my-skill)
}

// RepoInstall ensures the repo clone exists, adds the skill path to sparse
// checkout, validates SKILL.md, and creates the symlink.
func RepoInstall(opts RepoInstallOptions) (*CloneResult, error) {
	sparse := opts.Path != ""
	isNew := false

	// Step 1: Ensure repo clone exists
	if _, err := os.Stat(opts.RepoDir); os.IsNotExist(err) {
		isNew = true
		if err := ensureRepoClone(opts.RepoURL, opts.RepoDir, sparse); err != nil {
			return nil, err
		}
	}

	// Step 2: Add sparse path
	if sparse {
		if isNew {
			// First clone was --sparse, set the path
			if err := runGit(opts.RepoDir, "sparse-checkout", "set", opts.Path); err != nil {
				return nil, fmt.Errorf("sparse-checkout set failed: %w", err)
			}
		} else {
			// Repo already existed, add the new path (idempotent)
			if err := runGit(opts.RepoDir, "sparse-checkout", "add", opts.Path); err != nil {
				return nil, fmt.Errorf("sparse-checkout add failed: %w", err)
			}
		}
	}

	// Step 3: Resolve skill path in worktree
	var skillPath string
	if sparse {
		skillPath = filepath.Join(opts.RepoDir, opts.Path)
	} else {
		skillPath = opts.RepoDir
	}

	// Verify the path materialized
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		// Existing sparse clones can be stale (new skill path added upstream).
		// Try a fast-forward refresh once and re-apply sparse checkout.
		if sparse && !isNew {
			if err := refreshExistingClone(opts.RepoDir); err != nil {
				return nil, fmt.Errorf("skill path %s not found in repository after checkout (failed to refresh existing clone: %w)", opts.Path, err)
			}
			if err := runGit(opts.RepoDir, "sparse-checkout", "add", opts.Path); err != nil {
				return nil, fmt.Errorf("sparse-checkout add failed after refresh: %w", err)
			}
			if _, err := os.Stat(skillPath); os.IsNotExist(err) {
				return nil, fmt.Errorf("skill path %s not found in repository after checkout", opts.Path)
			}
		} else {
			return nil, fmt.Errorf("skill path %s not found in repository after checkout", opts.Path)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to validate skill path %s: %w", opts.Path, err)
	}

	// Step 4: Validate SKILL.md exists
	if err := ValidateSkill(skillPath); err != nil {
		return nil, err
	}

	// Step 5: Create symlink
	// Remove any existing item at the symlink path (symlink or dir)
	if info, err := os.Lstat(opts.SkillLink); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			os.Remove(opts.SkillLink)
		} else if info.IsDir() {
			os.RemoveAll(opts.SkillLink)
		} else {
			os.Remove(opts.SkillLink)
		}
	}

	if err := os.Symlink(skillPath, opts.SkillLink); err != nil {
		return nil, fmt.Errorf("failed to create symlink %s -> %s: %w", opts.SkillLink, skillPath, err)
	}

	// Step 6: Return result
	commit, err := getHeadCommit(opts.RepoDir)
	if err != nil {
		return nil, err
	}

	return &CloneResult{
		Commit: commit,
		Path:   skillPath,
	}, nil
}

// refreshExistingClone fast-forwards an existing clone to origin without
// destructive resets. This is used when sparse checkout paths were added
// upstream after the local clone was first created.
func refreshExistingClone(repoDir string) error {
	if err := runGit(repoDir, "fetch", "--depth", "1", "origin"); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}
	if err := runGit(repoDir, "merge", "--ff-only", "FETCH_HEAD"); err != nil {
		// If histories diverged/rebased, fall back to a hard reset only when
		// there are no local uncommitted changes to lose.
		modified, modErr := IsModified(repoDir)
		if modErr != nil {
			return fmt.Errorf("git merge --ff-only failed (%v), and failed to check local modifications: %w", err, modErr)
		}
		if modified {
			return fmt.Errorf("git merge --ff-only failed (%v), and repository has local modifications", err)
		}
		if resetErr := runGit(repoDir, "reset", "--hard", "FETCH_HEAD"); resetErr != nil {
			return fmt.Errorf("git merge --ff-only failed (%v), and git reset --hard failed: %w", err, resetErr)
		}
	}
	return nil
}

// ensureRepoClone clones a repository. If sparse is true, uses --sparse for
// cone-mode sparse checkout (only root files checked out initially).
func ensureRepoClone(repoURL, repoDir string, sparse bool) error {
	if err := os.MkdirAll(filepath.Dir(repoDir), 0755); err != nil {
		return fmt.Errorf("failed to create repos directory: %w", err)
	}

	if !sparse {
		// Full clone (the repo IS the skill)
		if err := runGit(".", "clone", repoURL, repoDir); err != nil {
			return fmt.Errorf("git clone failed: %w", err)
		}
		return nil
	}

	// Try sparse clone (cone mode)
	err := runGit(".", "clone", "--sparse", repoURL, repoDir)
	if err == nil {
		return nil
	}

	// Fallback: --no-checkout then init sparse-checkout manually
	os.RemoveAll(repoDir)
	if err := runGit(".", "clone", "--no-checkout", repoURL, repoDir); err != nil {
		return fmt.Errorf("git clone --no-checkout failed: %w", err)
	}
	if err := runGit(repoDir, "sparse-checkout", "init", "--cone"); err != nil {
		os.RemoveAll(repoDir)
		return fmt.Errorf("sparse-checkout init failed: %w", err)
	}
	if err := runGit(repoDir, "checkout"); err != nil {
		os.RemoveAll(repoDir)
		return fmt.Errorf("git checkout failed: %w", err)
	}

	return nil
}
