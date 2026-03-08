package history

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/backendsystems/nibble/internal/history/paths"
)

// ViewState represents the last selected item in the history view
type ViewState struct {
	SelectedPath  string         `json:"selected_path"`   // Path to currently selected item
	DetailCursors map[string]int `json:"detail_cursors"`  // Remembered host cursor per scan file path
}

var lastSavedPath = ""

func stateFile() (string, error) {
	stateDir, err := paths.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, ".view_state"), nil
}

func loadStateRaw() (ViewState, error) {
	f, err := stateFile()
	if err != nil {
		return ViewState{}, err
	}
	data, err := os.ReadFile(f)
	if err != nil {
		if os.IsNotExist(err) {
			return ViewState{}, nil
		}
		return ViewState{}, err
	}
	var state ViewState
	if err := json.Unmarshal(data, &state); err != nil {
		return ViewState{}, err
	}
	return state, nil
}

func writeState(state ViewState) error {
	f, err := stateFile()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f, append(data, '\n'), 0o644)
}

// SaveViewState persists the selected item path only if changed
func SaveViewState(selectedPath string) error {
	if selectedPath == lastSavedPath {
		return nil
	}
	state, _ := loadStateRaw()
	state.SelectedPath = selectedPath
	if err := writeState(state); err != nil {
		return err
	}
	lastSavedPath = selectedPath
	return nil
}

func toRelPath(absPath string) string {
	baseDir, err := paths.Dir()
	if err != nil {
		return absPath
	}
	rel, err := filepath.Rel(baseDir, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

func toAbsPath(relPath string) string {
	baseDir, err := paths.Dir()
	if err != nil {
		return relPath
	}
	return filepath.Join(baseDir, relPath)
}

// SaveDetailCursor persists the host cursor for a single scan file.
func SaveDetailCursor(historyPath string, cursor int) error {
	state, _ := loadStateRaw()
	if state.DetailCursors == nil {
		state.DetailCursors = make(map[string]int)
	}
	state.DetailCursors[toRelPath(historyPath)] = cursor
	return writeState(state)
}

// DeleteDetailCursors removes stored cursors for a batch of deleted scan files.
func DeleteDetailCursors(historyPaths []string) error {
	state, _ := loadStateRaw()
	if state.DetailCursors == nil {
		return nil
	}
	for _, p := range historyPaths {
		delete(state.DetailCursors, toRelPath(p))
	}
	return writeState(state)
}

// LoadViewState retrieves the last selected item path and detail cursors.
// Detail cursor keys are returned as absolute paths.
func LoadViewState() (string, map[string]int, error) {
	state, err := loadStateRaw()
	if err != nil {
		return "", nil, err
	}
	lastSavedPath = state.SelectedPath
	abs := make(map[string]int, len(state.DetailCursors))
	for rel, cur := range state.DetailCursors {
		abs[toAbsPath(rel)] = cur
	}
	return state.SelectedPath, abs, nil
}
