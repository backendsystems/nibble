package services

// Lookup returns information about the service commonly running on a port
// Returns nil if the port is not in the database
func Lookup(port int) *Info {
	// Check exact match first
	if info, ok := ports[port]; ok {
		return &info
	}

	// Check ranges
	for _, r := range ranges {
		if port >= r.start && port <= r.end {
			return &r.info
		}
	}

	return nil
}

// Name returns just the service name for a port, or empty string if unknown
func Name(port int) string {
	if info := Lookup(port); info != nil {
		return info.Name
	}
	return ""
}

// Description returns a human-readable description of the service on a port
// Returns "Unknown service" if the port is not in the database
func Description(port int) string {
	if info := Lookup(port); info != nil {
		return info.Description
	}
	return "Unknown service"
}

// Format returns a formatted string like "HTTPS - HTTP over TLS/SSL"
func Format(port int) string {
	info := Lookup(port)
	if info == nil {
		return ""
	}
	return info.Name + " - " + info.Description
}
