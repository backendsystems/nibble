package history

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/backendsystems/nibble/internal/history/paths"
)

// ViewState represents the last selected item in the history view
type ViewState struct {
	SelectedPath string `json:"selected_path"` // Path to currently selected item
}

var lastSavedPath = ""

// SaveViewState persists the selected item path only if changed
func SaveViewState(selectedPath string) error {
	// Only write if the path actually changed
	if selectedPath == lastSavedPath {
		return nil
	}

	stateDir, err := paths.Dir()
	if err != nil {
		return err
	}

	stateFile := filepath.Join(stateDir, ".view_state")
	state := ViewState{
		SelectedPath: selectedPath,
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(stateFile, append(data, '\n'), 0o644)
	if err == nil {
		lastSavedPath = selectedPath
	}
	return err
}

// LoadViewState retrieves the last selected item path
func LoadViewState() (string, error) {
	stateDir, err := paths.Dir()
	if err != nil {
		return "", err
	}

	stateFile := filepath.Join(stateDir, ".view_state")
	data, err := os.ReadFile(stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Return empty string if file doesn't exist yet
		}
		return "", err
	}

	var state ViewState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", err
	}

	lastSavedPath = state.SelectedPath
	return state.SelectedPath, nil
}
