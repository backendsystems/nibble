package wsl

import (
	"net"
	"strings"
)

// Interface represents a Windows network interface visible via WSL interop.
type Interface struct {
	Name   string
	Index  int
	Addrs  []net.Addr
	HWAddr net.HardwareAddr
}

// Interfaces returns Windows network interfaces by parsing ipconfig.exe /all output.
// Each interface is given a synthetic name like "win:Ethernet" to distinguish
// it from WSL-internal interfaces.
func Interfaces() []Interface {
	// ipconfig.exe is a native Win32 binary, much faster to start than PowerShell
	out, err := runWinCmd("ipconfig.exe", "/all")
	if err != nil {
		return nil
	}
	return parseIpconfig(out)
}

// parseIpconfig parses the output of `ipconfig.exe /all` into Interface entries.
func parseIpconfig(out string) []Interface {
	var result []Interface
	var cur *Interface
	var pendingIP net.IP
	var curMask net.IP

	commit := func() {
		if cur == nil {
			return
		}
		// flush any IP that never got its mask
		if pendingIP != nil {
			cur.Addrs = append(cur.Addrs, &net.IPNet{IP: pendingIP, Mask: net.CIDRMask(32, 32)})
			pendingIP = nil
		}
		curMask = nil
		if len(cur.Addrs) > 0 {
			result = append(result, *cur)
		}
		cur = nil
	}

	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimRight(line, "\r")

		// Adapter section header — not indented, ends with ":"
		// e.g. "Ethernet adapter Ethernet:" or "Wireless LAN adapter Wi-Fi:"
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			commit()
			name := strings.TrimSuffix(strings.TrimSpace(line), ":")
			lower := strings.ToLower(name)
			if strings.Contains(lower, "loopback") || strings.Contains(lower, "tunnel") || strings.Contains(lower, "teredo") || strings.Contains(lower, "vethernet") {
				continue
			}
			cur = &Interface{Name: "win:" + adapterShortName(name)}
			continue
		}

		if cur == nil {
			continue
		}

		key, val, ok := splitIpconfigField(line)
		if !ok {
			continue
		}

		switch {
		case strings.Contains(key, "Physical Address"):
			mac, err := net.ParseMAC(strings.ReplaceAll(val, "-", ":"))
			if err == nil {
				cur.HWAddr = mac
			}

		case strings.Contains(key, "IPv4 Address"):
			// value may be "192.168.1.5(Preferred)" or "192.168.1.5"
			ipStr := strings.TrimSuffix(strings.TrimSuffix(val, "(Preferred)"), "(Duplicate)")
			parsed := net.ParseIP(ipStr).To4()
			if parsed == nil || parsed.IsLinkLocalUnicast() || parsed.IsLoopback() {
				continue
			}
			// flush any previous pending IP that never got a mask
			if pendingIP != nil {
				cur.Addrs = append(cur.Addrs, &net.IPNet{IP: pendingIP, Mask: net.CIDRMask(32, 32)})
			}
			if curMask != nil {
				cur.Addrs = append(cur.Addrs, &net.IPNet{IP: parsed, Mask: net.IPMask(curMask.To4())})
				pendingIP = nil
				curMask = nil
			} else {
				pendingIP = parsed
			}

		case strings.Contains(key, "Subnet Mask"):
			mask := net.ParseIP(val).To4()
			if mask == nil {
				continue
			}
			if pendingIP != nil {
				cur.Addrs = append(cur.Addrs, &net.IPNet{IP: pendingIP, Mask: net.IPMask(mask)})
				pendingIP = nil
			} else {
				curMask = mask
			}
		}
	}
	commit()
	return result
}

// splitIpconfigField splits a line like "   Physical Address. . . . . : AA-BB-CC-DD-EE-FF"
// into key="Physical Address" and val="AA-BB-CC-DD-EE-FF".
func splitIpconfigField(line string) (key, val string, ok bool) {
	idx := strings.Index(line, " : ")
	if idx < 0 {
		return
	}
	key = strings.TrimRight(strings.TrimSpace(line[:idx]), ". ")
	val = strings.TrimSpace(line[idx+3:])
	ok = true
	return
}

// adapterShortName extracts the user-visible name from an ipconfig adapter header.
// "Ethernet adapter Ethernet" -> "Ethernet"
// "Wireless LAN adapter Wi-Fi" -> "Wi-Fi"
func adapterShortName(header string) string {
	for _, prefix := range []string{
		"Ethernet adapter ",
		"Wireless LAN adapter ",
		"PPP adapter ",
		"Tunnel adapter ",
	} {
		if strings.HasPrefix(header, prefix) {
			return strings.TrimPrefix(header, prefix)
		}
	}
	return header
}

