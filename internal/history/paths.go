package history

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HistoryDir returns the base history directory path
func HistoryDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "nibble", "history"), nil
}

// FormatCIDRPath converts a CIDR like "192.168.1.0/24" to "192.168.1.0_24"
func FormatCIDRPath(cidr string) string {
	return strings.ReplaceAll(cidr, "/", "_")
}

// ScanPath builds the full path for a scan file
func ScanPath(interfaceName, cidr string, timestamp time.Time) (string, error) {
	base, err := HistoryDir()
	if err != nil {
		return "", err
	}

	cidrPath := FormatCIDRPath(cidr)
	filename := fmt.Sprintf("scan_%s.json", timestamp.Format("20060102_150405"))

	return filepath.Join(base, interfaceName, cidrPath, filename), nil
}

// ScanDir returns the directory path for a specific interface/CIDR combination
func ScanDir(interfaceName, cidr string) (string, error) {
	base, err := HistoryDir()
	if err != nil {
		return "", err
	}

	cidrPath := FormatCIDRPath(cidr)
	return filepath.Join(base, interfaceName, cidrPath), nil
}
