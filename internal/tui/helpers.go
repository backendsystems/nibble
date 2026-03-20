package tui

func scanViewWidth(windowW int) int {
	maxWidth := 72
	if windowW > 8 {
		maxWidth = windowW - 4
	}
	return maxWidth
}
