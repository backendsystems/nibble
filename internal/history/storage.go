package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Save writes a scan history to disk
func Save(history ScanHistory) error {
	path, err := ScanPath(
		history.ScanMetadata.InterfaceName,
		history.ScanMetadata.TargetCIDR,
		history.ScanMetadata.Created,
	)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

// Load reads a scan history from disk
func Load(path string) (ScanHistory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ScanHistory{}, err
	}

	var history ScanHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return ScanHistory{}, err
	}

	return history, nil
}

// Update updates an existing scan history file
func Update(path string, history ScanHistory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

// Delete removes a scan history file
func Delete(path string) error {
	return os.Remove(path)
}

// ListAll returns all scan history files sorted by creation time (newest first)
func ListAll() ([]ScanHistory, []string, error) {
	base, err := HistoryDir()
	if err != nil {
		return nil, nil, err
	}

	var histories []ScanHistory
	var paths []string

	// Walk through all subdirectories
	err = filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".json" {
			return nil
		}

		history, err := Load(path)
		if err != nil {
			return nil // Skip invalid files
		}

		histories = append(histories, history)
		paths = append(paths, path)
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Sort by created time, newest first
	sort.Slice(histories, func(i, j int) bool {
		return histories[i].ScanMetadata.Created.After(histories[j].ScanMetadata.Created)
	})

	// Sort paths in the same order
	sort.Slice(paths, func(i, j int) bool {
		hi, _ := Load(paths[i])
		hj, _ := Load(paths[j])
		return hi.ScanMetadata.Created.After(hj.ScanMetadata.Created)
	})

	return histories, paths, nil
}

// UpdateHostInScan updates a specific host in an existing scan history
// Completely replaces the host data since rescans assume everything can change
func UpdateHostInScan(path string, hostIP string, newHost HostResult) error {
	history, err := Load(path)
	if err != nil {
		return err
	}

	// Update the scan's updated timestamp
	history.ScanMetadata.Updated = time.Now()

	// Find and replace the host completely
	found := false
	for i, host := range history.ScanResults.Hosts {
		if host.IP == hostIP {
			history.ScanResults.Hosts[i] = newHost
			found = true
			break
		}
	}

	// If host not found, add it
	if !found {
		history.ScanResults.Hosts = append(history.ScanResults.Hosts, newHost)
		history.ScanResults.HostsFound++
	}

	// Recalculate total ports found
	totalPorts := 0
	for _, host := range history.ScanResults.Hosts {
		totalPorts += len(host.Ports)
	}
	history.ScanResults.PortsFound = totalPorts

	return Update(path, history)
}
