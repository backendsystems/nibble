package history

import (
	"time"

	"github.com/backendsystems/nibble/internal/history/paths"
)

// HistoryDir returns the base history directory path
func HistoryDir() (string, error) {
	return paths.Dir()
}

// FormatCIDRPath converts a CIDR like "192.168.1.0/24" to "192.168.1.0_24"
func FormatCIDRPath(cidr string) string {
	return paths.FormatCIDR(cidr)
}

// ScanPath builds the full path for a scan file
func ScanPath(interfaceName, cidr string, timestamp time.Time) (string, error) {
	return paths.Scan(interfaceName, cidr, timestamp)
}

// ScanDir returns the directory path for a specific interface/CIDR combination
func ScanDir(interfaceName, cidr string) (string, error) {
	return paths.ScanDir(interfaceName, cidr)
}
