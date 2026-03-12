package historydetailview

// ScrollToSelected adjusts the viewport offset so the selected host is visible,
// showing as many of its ports as possible without pushing the host IP off-screen.
// During scanning it follows the scanning host, keeping new ports visible at the bottom.
func (m Model) ScrollToSelected() Model {
	hosts := m.History.ScanResults.Hosts
	if len(hosts) == 0 {
		return m
	}

	if m.Scanning && m.ScanningHostIdx >= 0 && m.ScanningHostIdx < len(hosts) {
		// Follow the scanning host: keep its latest port visible at the bottom.
		scanStart := hostLineOffsetFor(m, m.ScanningHostIdx)
		scanHost := hosts[m.ScanningHostIdx]
		scanEnd := scanStart + len(scanHost.Ports)
		bottom := m.Viewport.YOffset + m.Viewport.Height - 1
		if scanEnd > bottom {
			offset := scanEnd - m.Viewport.Height + 1
			if offset > scanStart {
				offset = scanStart
			}
			m.Viewport.YOffset = offset
		}
		if m.Viewport.YOffset < 0 {
			m.Viewport.YOffset = 0
		}
		return m
	}

	if m.Cursor < 0 || m.Cursor >= len(hosts) {
		return m
	}

	hostStart := hostLineOffsetFor(m, m.Cursor)
	selectedHost := hosts[m.Cursor]
	hostEnd := hostStart + len(selectedHost.Ports)

	top := m.Viewport.YOffset
	bottom := m.Viewport.YOffset + m.Viewport.Height - 1

	if hostStart < top {
		// Host IP is above viewport: scroll up to show it.
		// For the first host, scroll to 0 so the metadata lines above are visible too.
		if m.Cursor == 0 {
			m.Viewport.YOffset = 0
		} else {
			m.Viewport.YOffset = hostStart
		}
	} else if hostStart > bottom {
		// Host IP is below viewport: scroll down just enough to show host + ports.
		offset := hostEnd - m.Viewport.Height + 1
		if offset > hostStart {
			offset = hostStart
		}
		if offset < 0 {
			offset = 0
		}
		m.Viewport.YOffset = offset
	} else if hostEnd > bottom {
		// Host IP is visible but ports extend below: scroll down to show ports,
		// but never push the host IP itself off the top.
		offset := hostEnd - m.Viewport.Height + 1
		if offset > hostStart {
			offset = hostStart
		}
		if offset <= top {
			// Already showing as much as possible without losing the host line.
			return m
		}
		m.Viewport.YOffset = offset
	}

	if m.Viewport.YOffset < 0 {
		m.Viewport.YOffset = 0
	}

	return m
}

// hostLineOffsetFor returns the viewport content line index of the host at idx.
func hostLineOffsetFor(m Model, idx int) int {
	offset := 0

	// Metadata lines rendered before the host list.
	if m.History.ScanResults.PortsFound > 0 {
		offset++
	}
	offset++ // Created/Updated line

	for i := 0; i < idx; i++ {
		host := m.History.ScanResults.Hosts[i]
		offset++                  // Host line
		offset += len(host.Ports) // Port lines
	}

	return offset
}
