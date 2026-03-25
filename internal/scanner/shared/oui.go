package shared

import (
	"net"

	"github.com/endobit/oui"
)

// VendorFromMac returns the hardware manufacturer from a MAC address OUI prefix
// Returns "" if no vendor is found
func VendorFromMac(mac string) string {
	if vendor := oui.Vendor(mac); vendor != "" {
		return vendor
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
