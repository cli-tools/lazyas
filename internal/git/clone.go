package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneOptions specifies options for cloning
type CloneOptions struct {
	Repo      string
	Path      string // subdirectory within repo (optional)
	Tag       string // version tag or branch
	TargetDir string // where to clone to
}

// CloneResult contains the result of a clone operation
type CloneResult struct {
	Commit string
	Path   string
}

// Clone clones a repository or subdirectory
func Clone(opts CloneOptions) (*CloneResult, error) {
	// If no subdirectory, do a simple clone
	if opts.Path == "" {
		return cloneFullRepo(opts)
	}

	// Use sparse checkout for subdirectory
	return cloneSparse(opts)
}

func cloneFullRepo(opts CloneOptions) (*CloneResult, error) {
	args := []string{"clone", "--depth", "1"}

	if opts.Tag != "" {
		args = append(args, "--branch", opts.Tag)
	}

	args = append(args, opts.Repo, opts.TargetDir)

	cmd := exec.Command("git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git clone failed: %w\n%s", err, stderr.String())
	}

	commit, err := getHeadCommit(opts.TargetDir)
	if err != nil {
		return nil, err
	}

	return &CloneResult{
		Commit: commit,
		Path:   opts.TargetDir,
	}, nil
}

func cloneSparse(opts CloneOptions) (*CloneResult, error) {
	// Use git sparse-checkout to clone only the subdirectory
	// but preserve the .git directory for tracking changes

	// Initialize empty repo at target
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target dir: %w", err)
	}

	// Initialize repo
	if err := runGit(opts.TargetDir, "init"); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git init failed: %w", err)
	}

	// Add remote
	if err := runGit(opts.TargetDir, "remote", "add", "origin", opts.Repo); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git remote add failed: %w", err)
	}

	// Enable sparse checkout with cone mode for better performance
	if err := runGit(opts.TargetDir, "sparse-checkout", "init", "--cone"); err != nil {
		// Fallback to legacy sparse checkout if cone mode not supported
		if err := runGit(opts.TargetDir, "config", "core.sparseCheckout", "true"); err != nil {
			os.RemoveAll(opts.TargetDir)
			return nil, fmt.Errorf("failed to enable sparse checkout: %w", err)
		}
	}

	// Set sparse checkout path
	if err := runGit(opts.TargetDir, "sparse-checkout", "set", opts.Path); err != nil {
		// Fallback to manual sparse-checkout file
		sparseFile := filepath.Join(opts.TargetDir, ".git", "info", "sparse-checkout")
		if err := os.MkdirAll(filepath.Dir(sparseFile), 0755); err != nil {
			os.RemoveAll(opts.TargetDir)
			return nil, fmt.Errorf("failed to create sparse-checkout dir: %w", err)
		}
		sparsePath := strings.TrimSuffix(opts.Path, "/") + "/*"
		if err := os.WriteFile(sparseFile, []byte(sparsePath+"\n"), 0644); err != nil {
			os.RemoveAll(opts.TargetDir)
			return nil, fmt.Errorf("failed to write sparse-checkout: %w", err)
		}
	}

	// Fetch with depth 1
	ref := "HEAD"
	if opts.Tag != "" {
		ref = opts.Tag
	}
	if err := runGit(opts.TargetDir, "fetch", "--depth", "1", "origin", ref); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git fetch failed: %w", err)
	}

	// Checkout and create tracking branch
	if err := runGit(opts.TargetDir, "checkout", "FETCH_HEAD"); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git checkout failed: %w", err)
	}

	commit, err := getHeadCommit(opts.TargetDir)
	if err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, err
	}

	// Check if the skill path exists in the checkout
	skillPath := filepath.Join(opts.TargetDir, opts.Path)
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("skill path %s not found in repository", opts.Path)
	}

	// Relocate files from the nested subdirectory to the target root so that
	// SKILL.md (and everything else) lives at targetDir/ instead of
	// targetDir/<opts.Path>/.
	entries, err := os.ReadDir(skillPath)
	if err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("failed to read skill path: %w", err)
	}
	for _, e := range entries {
		src := filepath.Join(skillPath, e.Name())
		dst := filepath.Join(opts.TargetDir, e.Name())
		if err := os.Rename(src, dst); err != nil {
			os.RemoveAll(opts.TargetDir)
			return nil, fmt.Errorf("failed to relocate %s: %w", e.Name(), err)
		}
	}

	// Remove the now-empty intermediate directory tree.
	// topLevel is the first path component (e.g. "skills" from "skills/foo").
	topLevel := strings.SplitN(opts.Path, "/", 2)[0]
	os.RemoveAll(filepath.Join(opts.TargetDir, topLevel))

	// Remove the old .git whose index tracks the original nested layout.
	os.RemoveAll(filepath.Join(opts.TargetDir, ".git"))

	// Re-initialise a fresh repo so modification tracking works against the
	// relocated file layout.
	if err := runGit(opts.TargetDir, "init"); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git re-init failed: %w", err)
	}
	if err := runGit(opts.TargetDir, "remote", "add", "origin", opts.Repo); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git remote add failed after relocate: %w", err)
	}
	// Mark this as a relocated sparse skill so Update() knows how to handle it.
	runGit(opts.TargetDir, "config", "lazyas.path", opts.Path)

	// Baseline commit for modification tracking.
	if err := runGit(opts.TargetDir, "add", "-A"); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git add failed: %w", err)
	}
	if err := runGit(opts.TargetDir, "commit", "-m", "lazyas install"); err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, fmt.Errorf("git commit failed: %w", err)
	}

	// Re-read commit hash from the fresh repo.
	commit, err = getHeadCommit(opts.TargetDir)
	if err != nil {
		os.RemoveAll(opts.TargetDir)
		return nil, err
	}

	return &CloneResult{
		Commit: commit,
		Path:   opts.TargetDir,
	}, nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		errMsg := stderr.String()
		if errMsg != "" {
			return fmt.Errorf("%w\n%s", err, errMsg)
		}
		return err
	}
	return nil
}

func getGitConfig(dir, key string) string {
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func getHeadCommit(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo checks if the path is a git repository
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsModified checks if a git repo has local modifications
func IsModified(path string) (bool, error) {
	if !IsGitRepo(path) {
		return false, nil // Not a git repo, can't be modified
	}

	// Check for uncommitted changes (staged or unstaged)
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status failed: %w", err)
	}

	return len(strings.TrimSpace(string(out))) > 0, nil
}

// GetModifiedFiles returns list of modified files in a git repo
func GetModifiedFiles(path string) ([]string, error) {
	if !IsGitRepo(path) {
		return nil, nil
	}

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if len(line) > 3 {
			files = append(files, line[3:]) // Skip status prefix
		}
	}
	return files, nil
}

// GetDiff returns the diff of local changes
func GetDiff(path string) (string, error) {
	if !IsGitRepo(path) {
		return "", nil
	}

	cmd := exec.Command("git", "diff", "HEAD")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

// Update pulls the latest changes for a skill
// Returns error if there are local modifications (to prevent losing changes)
func Update(skillPath, tag string) (*CloneResult, error) {
	// Check for local modifications first
	modified, err := IsModified(skillPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check for modifications: %w", err)
	}
	if modified {
		return nil, fmt.Errorf("skill has local modifications; commit or discard changes before updating")
	}

	// If this is a relocated sparse skill, re-clone from scratch so we
	// don't restore the original nested layout via reset --hard.
	if sparsePath := getGitConfig(skillPath, "lazyas.path"); sparsePath != "" {
		repo := getGitConfig(skillPath, "remote.origin.url")
		if repo == "" {
			return nil, fmt.Errorf("relocated sparse skill has no remote.origin.url")
		}
		os.RemoveAll(skillPath)
		return Clone(CloneOptions{
			Repo:      repo,
			Path:      sparsePath,
			Tag:       tag,
			TargetDir: skillPath,
		})
	}

	// For shallow clones, we need to fetch and reset
	if tag != "" {
		if err := runGit(skillPath, "fetch", "--depth", "1", "origin", tag); err != nil {
			return nil, fmt.Errorf("git fetch failed: %w", err)
		}
		if err := runGit(skillPath, "reset", "--hard", "FETCH_HEAD"); err != nil {
			return nil, fmt.Errorf("git reset failed: %w", err)
		}
	} else {
		if err := runGit(skillPath, "fetch", "--depth", "1", "origin"); err != nil {
			return nil, fmt.Errorf("git fetch failed: %w", err)
		}
		if err := runGit(skillPath, "reset", "--hard", "FETCH_HEAD"); err != nil {
			return nil, fmt.Errorf("git reset failed: %w", err)
		}
	}

	commit, err := getHeadCommit(skillPath)
	if err != nil {
		return nil, err
	}

	return &CloneResult{
		Commit: commit,
		Path:   skillPath,
	}, nil
}

// ResetChanges discards all local modifications
func ResetChanges(path string) error {
	if !IsGitRepo(path) {
		return fmt.Errorf("not a git repository")
	}

	if err := runGit(path, "checkout", "."); err != nil {
		return fmt.Errorf("git checkout failed: %w", err)
	}
	if err := runGit(path, "clean", "-fd"); err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}
	return nil
}
