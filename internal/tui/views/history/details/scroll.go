package historydetailview

// scrollToSelected adjusts the viewport offset so the selected host is visible,
// showing as many of its ports as possible without pushing the host IP off-screen.
func (m Model) scrollToSelected() Model {
	hosts := m.History.ScanResults.Hosts
	if len(hosts) == 0 || m.Cursor < 0 || m.Cursor >= len(hosts) {
		return m
	}

	hostStart := hostLineOffset(m)
	selectedHost := hosts[m.Cursor]
	hostEnd := hostStart + len(selectedHost.Ports)

	top := m.Viewport.YOffset
	bottom := m.Viewport.YOffset + m.Viewport.Height - 1

	if hostStart < top {
		// Selection scrolled above viewport: bring host line to top.
		m.Viewport.YOffset = hostStart
	} else if hostStart > bottom {
		// Selection scrolled below viewport: show host IP + as many ports as fit.
		// Target hostEnd at bottom; cap at hostStart so host IP never goes off-screen.
		offset := hostEnd - m.Viewport.Height + 1
		if offset > hostStart {
			offset = hostStart
		}
		m.Viewport.YOffset = offset
	}

	if m.Viewport.YOffset < 0 {
		m.Viewport.YOffset = 0
	}

	return m
}

// hostLineOffset returns the viewport content line index of the selected host.
func hostLineOffset(m Model) int {
	offset := 0

	// Metadata lines rendered before the host list.
	if m.History.ScanResults.PortsFound > 0 {
		offset++
	}
	offset++ // Created/Updated line

	for i := 0; i < m.Cursor; i++ {
		host := m.History.ScanResults.Hosts[i]
		offset++                  // Host line
		offset += len(host.Ports) // Port lines
	}

	return offset
}
