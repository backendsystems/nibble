package history

import (
	"github.com/backendsystems/nibble/internal/history/delete"
)

// Delete removes a scan history file, with path escape guards
func Delete(path string) error {
	return delete.Delete(path)
}

// DeleteDir removes all .json files in a directory, with path escape guards
func DeleteDir(dirPath string) error {
	return delete.DeleteDir(dirPath)
}
