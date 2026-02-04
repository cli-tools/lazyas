package symlink

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"lazyas/internal/config"
)

// LinkStatus represents the status of a backend symlink
type LinkStatus struct {
	Backend     config.Backend
	Linked      bool   // Is the symlink properly configured?
	Exists      bool   // Does the target path exist?
	HasFiles    bool   // Does the target have existing files?
	IsSymlink   bool   // Is the target already a symlink?
	SymlinkDest string // Where does the symlink point (if it's a symlink)
	Error       error  // Any error encountered
}

// CheckBackendLinks checks the symlink status for all backends
func CheckBackendLinks(backends []config.Backend, centralDir string) []LinkStatus {
	results := make([]LinkStatus, len(backends))

	for i, backend := range backends {
		results[i] = checkSingleBackend(backend, centralDir)
	}

	return results
}

// checkSingleBackend checks the symlink status for a single backend
func checkSingleBackend(backend config.Backend, centralDir string) LinkStatus {
	status := LinkStatus{
		Backend: backend,
	}

	// Expand the backend path
	backendPath, err := config.ExpandPath(backend.Path)
	if err != nil {
		status.Error = fmt.Errorf("failed to expand path: %w", err)
		return status
	}

	// Check if path exists
	info, err := os.Lstat(backendPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Path doesn't exist - not linked, no files
			return status
		}
		status.Error = fmt.Errorf("failed to stat path: %w", err)
		return status
	}

	status.Exists = true

	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		status.IsSymlink = true

		// Read the symlink target
		target, err := os.Readlink(backendPath)
		if err != nil {
			status.Error = fmt.Errorf("failed to read symlink: %w", err)
			return status
		}
		status.SymlinkDest = target

		// Make target absolute for comparison
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(backendPath), target)
		}
		target = filepath.Clean(target)

		// Check if symlink points to our central directory
		status.Linked = target == centralDir
		return status
	}

	// It's a regular directory - check if it has files
	entries, err := os.ReadDir(backendPath)
	if err != nil {
		status.Error = fmt.Errorf("failed to read directory: %w", err)
		return status
	}
	status.HasFiles = len(entries) > 0

	return status
}

// CreateLink creates a symlink from the backend path to the central directory
func CreateLink(backend config.Backend, centralDir string) error {
	backendPath, err := config.ExpandPath(backend.Path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	// Ensure the parent directory exists
	parentDir := filepath.Dir(backendPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Ensure central directory exists
	if err := os.MkdirAll(centralDir, 0755); err != nil {
		return fmt.Errorf("failed to create central directory: %w", err)
	}

	// Create the symlink
	if runtime.GOOS == "windows" {
		return createWindowsLink(backendPath, centralDir)
	}

	return os.Symlink(centralDir, backendPath)
}

// createWindowsLink creates a directory junction on Windows
func createWindowsLink(linkPath, targetPath string) error {
	// On Windows, we use mklink /J for directory junctions
	// This doesn't require admin privileges unlike /D for symlinks
	// For now, use standard symlink which requires developer mode or admin
	return os.Symlink(targetPath, linkPath)
}

// RemoveLink removes a symlink (but not a real directory)
func RemoveLink(backend config.Backend) error {
	backendPath, err := config.ExpandPath(backend.Path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	info, err := os.Lstat(backendPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Already doesn't exist
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}

	// Only remove if it's a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return fmt.Errorf("path is not a symlink, refusing to remove")
	}

	return os.Remove(backendPath)
}

// MigrateExistingDir moves files from an existing backend directory to the central directory
// and creates a symlink in place of the original directory
func MigrateExistingDir(backend config.Backend, centralDir string) error {
	backendPath, err := config.ExpandPath(backend.Path)
	if err != nil {
		return fmt.Errorf("failed to expand path: %w", err)
	}

	// Check that source exists and is a real directory (not a symlink)
	info, err := os.Lstat(backendPath)
	if err != nil {
		return fmt.Errorf("failed to stat backend path: %w", err)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("backend path is already a symlink")
	}

	if !info.IsDir() {
		return fmt.Errorf("backend path is not a directory")
	}

	// Ensure central directory exists
	if err := os.MkdirAll(centralDir, 0755); err != nil {
		return fmt.Errorf("failed to create central directory: %w", err)
	}

	// Move all contents from backend dir to central dir
	entries, err := os.ReadDir(backendPath)
	if err != nil {
		return fmt.Errorf("failed to read backend directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(backendPath, entry.Name())
		dstPath := filepath.Join(centralDir, entry.Name())

		// Check if destination already exists
		if _, err := os.Stat(dstPath); err == nil {
			// Skip if already exists - user can resolve conflicts manually
			continue
		}

		// Move the file/directory
		if err := os.Rename(srcPath, dstPath); err != nil {
			// If rename fails (cross-device), try copy+delete
			if err := copyRecursive(srcPath, dstPath); err != nil {
				return fmt.Errorf("failed to move %s: %w", entry.Name(), err)
			}
			os.RemoveAll(srcPath)
		}
	}

	// Remove the now-empty directory
	if err := os.Remove(backendPath); err != nil {
		return fmt.Errorf("failed to remove original directory: %w", err)
	}

	// Create symlink
	return CreateLink(backend, centralDir)
}

// copyRecursive copies a file or directory recursively
func copyRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, data, info.Mode())
}

// HasUnlinkedBackends returns true if any backend is not linked
func HasUnlinkedBackends(statuses []LinkStatus) bool {
	for _, s := range statuses {
		if !s.Linked && s.Error == nil {
			return true
		}
	}
	return false
}

// GetUnlinkedBackends returns backends that are not linked
func GetUnlinkedBackends(statuses []LinkStatus) []LinkStatus {
	var unlinked []LinkStatus
	for _, s := range statuses {
		if !s.Linked && s.Error == nil {
			unlinked = append(unlinked, s)
		}
	}
	return unlinked
}
