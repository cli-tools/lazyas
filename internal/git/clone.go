package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneResult contains the result of a clone operation
type CloneResult struct {
	Commit string
	Path   string
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

func getHeadCommit(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsGitRepo checks if the path is a git repository.
// Accepts both .git directories and .git files (gitlinks used by worktrees/submodules).
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// IsModified checks if a git repo has local modifications
func IsModified(path string) (bool, error) {
	if !IsGitRepo(path) {
		return false, nil // Not a git repo, can't be modified
	}

	// Check for uncommitted changes (staged or unstaged), scoped to current dir
	cmd := exec.Command("git", "status", "--porcelain", "--", ".")
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

	cmd := exec.Command("git", "status", "--porcelain", "--", ".")
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

	cmd := exec.Command("git", "diff", "HEAD", "--", ".")
	cmd.Dir = path
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(out), nil
}

// RemoteHEAD returns the HEAD commit of the remote origin without modifying
// local state. Requires a single network round-trip (git ls-remote).
func RemoteHEAD(repoDir string) (string, error) {
	cmd := exec.Command("git", "ls-remote", "origin", "HEAD")
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git ls-remote failed: %w", err)
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return "", fmt.Errorf("git ls-remote returned empty output")
	}
	return fields[0], nil
}

// IsRepoOutdated checks whether the local HEAD differs from the remote HEAD.
// Returns false (not outdated) on any error so callers can silently ignore failures.
func IsRepoOutdated(repoDir string) (bool, error) {
	localHead, err := getHeadCommit(repoDir)
	if err != nil {
		return false, err
	}
	remoteHead, err := RemoteHEAD(repoDir)
	if err != nil {
		return false, err
	}
	return localHead != remoteHead, nil
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

	// Fetch and reset to the target tag (or default branch)
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
