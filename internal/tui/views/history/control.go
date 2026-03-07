package historyview

import (
	detailsview "github.com/backendsystems/nibble/internal/tui/views/history/details"
	historytree "github.com/backendsystems/nibble/internal/tui/views/history/tree"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Init() tea.Cmd {
	return loadTreeCmd()
}

func (m Model) Update(msg tea.Msg) UpdateResult {
	result := UpdateResult{Model: m}

	// Route detail view messages
	if m.Mode == ViewDetail {
		detailResult := m.Details.Update(msg)
		result.Model.Details = detailResult.Model
		syncScanNode(result.Model.Tree, detailResult.Model.HistoryPath, detailResult.Model.History)
		result.Cmd = detailResult.Cmd
		if detailResult.Deleted {
			tree, selectedPath, _ := historytree.Build()
			result.Model.Tree = tree
			result.Model.FlatList = historytree.Flatten(tree)
			if selectedPath != "" {
				result.Model.Cursor = historytree.FindCursorByPath(result.Model.FlatList, selectedPath)
			}
			if result.Model.Cursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
				result.Model.Cursor = len(result.Model.FlatList) - 1
			}
			result.Model.Mode = ViewList
			result.Model.Details = detailsview.Model{}
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		} else if detailResult.Quit {
			result.Model.Mode = ViewList
			result.Model.Details = detailsview.Model{}
			saveViewState(result.Model.FlatList, result.Model.Cursor)
		}
		if detailResult.ScanAllPorts {
			result.ScanAllPorts = true
			result.SelectedHostIP = detailResult.SelectedHostIP
			result.ScanHistoryPath = detailResult.ScanHistoryPath
		}
		return result
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		result = handleKeyMsg(m, msg)
	case treeLoadedMsg:
		result.Model.Tree = msg.tree
		result.Model.FlatList = historytree.Flatten(result.Model.Tree)
		if msg.selectedPath != "" {
			result.Model.Cursor = historytree.FindCursorByPath(result.Model.FlatList, msg.selectedPath)
		}
		if result.Model.Cursor >= len(result.Model.FlatList) {
			result.Model.Cursor = 0
		}
		// Fire count loads for any network nodes already expanded by state restore
		result.Cmd = historytree.LoadCountsForExpandedNodes(result.Model.Tree)
	case historytree.ScanCountLoadedMsg:
		historytree.ApplyScanCountLoadedMsg(result.Model.FlatList, msg)
	default:
		var cmd tea.Cmd
		result.Model.Viewport, cmd = m.Viewport.Update(msg)
		_ = cmd
	}

	if result.Model.WindowW > 0 {
		oldListHeight := result.Model.Viewport.Height
		result.Model = result.Model.SetListViewportSize(result.Model.WindowW, result.Model.WindowH)
		if oldListHeight != result.Model.Viewport.Height {
			result.Model.Viewport.YOffset = 0
		}
		result.Model.Details.WindowW = result.Model.WindowW
		result.Model.Details.WindowH = result.Model.WindowH
	}

	if !result.Quit {
		result.Model = updateViewportContent(result.Model)
	}

	return result
}
