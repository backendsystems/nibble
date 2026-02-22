package shared

import (
	"net"
	"strings"

	"github.com/endobit/oui"
)

// VendorFromMac returns the hardware manufacturer from a MAC address OUI prefix
// Falls back to the uppercase MAC string if no vendor is found
func VendorFromMac(mac string) string {
	if vendor := oui.Vendor(mac); vendor != "" {
		return vendor
	}
	if mac != "" {
		return strings.ToUpper(mac)
	}
	return ""
}

// NormalizeMAC parses and normalizes a MAC address string into xx:xx:xx:xx:xx:xx form
// Returns "" for any invalid input
func NormalizeMAC(mac string) string {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) != 6 {
		return ""
	}
	return hw.String()
}
