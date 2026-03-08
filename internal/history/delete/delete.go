package delete

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/backendsystems/nibble/internal/history/paths"
)

// Delete removes a scan history file with escape guards
func Delete(path string) error {
	base, err := paths.Dir()
	if err != nil {
		return err
	}
	return SafeDelete(path, base)
}

// DeleteDir removes all .json files in a directory with escape guards
func DeleteDir(dirPath string) error {
	base, err := paths.Dir()
	if err != nil {
		return err
	}
	return SafeDeleteDir(dirPath, base)
}

// SafeDelete removes a file, ensuring it's within an allowed base directory
func SafeDelete(path, baseDir string) error {
	// Verify path is within base directory
	absPath, _ := filepath.Abs(path)
	absBase, _ := filepath.Abs(baseDir)
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		return os.ErrPermission
	}

	return os.Remove(absPath)
}

// SafeDeleteDir removes all .json files in a directory within a base directory
func SafeDeleteDir(dirPath, baseDir string) error {
	// Verify directory is within base directory
	absPath, _ := filepath.Abs(dirPath)
	absBase, _ := filepath.Abs(baseDir)
	if !strings.HasPrefix(absPath, absBase+string(filepath.Separator)) {
		return os.ErrPermission
	}

	// Walk and delete only .json files
	filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && filepath.Ext(path) == ".json" {
			os.Remove(path)
		}
		return nil
	})

	return nil
}
