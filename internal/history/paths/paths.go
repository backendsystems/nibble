package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Dir returns the base history directory path
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "nibble", "history"), nil
}

// FormatCIDR converts a CIDR like "192.168.1.0/24" to "192.168.1.0_24"
func FormatCIDR(cidr string) string {
	return strings.ReplaceAll(cidr, "/", "_")
}

// Scan builds the full path for a scan file
func Scan(interfaceName, cidr string, timestamp time.Time) (string, error) {
	base, err := Dir()
	if err != nil {
		return "", err
	}

	cidrPath := FormatCIDR(cidr)
	filename := fmt.Sprintf("scan_%s.json", timestamp.Format("20060102_150405"))

	return filepath.Join(base, interfaceName, cidrPath, filename), nil
}

// ScanDir returns the directory path for a specific interface/CIDR combination
func ScanDir(interfaceName, cidr string) (string, error) {
	base, err := Dir()
	if err != nil {
		return "", err
	}

	cidrPath := FormatCIDR(cidr)
	return filepath.Join(base, interfaceName, cidrPath), nil
}
