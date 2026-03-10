package historyview

import (
	"github.com/backendsystems/nibble/internal/history"
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
		if mouseMsg, ok := msg.(tea.MouseMsg); ok {
			detailResult := m.Details.HandleMouse(mouseMsg)
			result.Model.Details = detailResult.Model
			syncScanNode(result.Model.Tree, detailResult.Model.HistoryPath, detailResult.Model.History)
			result.Cmd = detailResult.Cmd
			applyDetailResult(&result, detailResult)
			return result
		}
		detailResult := m.Details.Update(msg)
		result.Model.Details = detailResult.Model
		syncScanNode(result.Model.Tree, detailResult.Model.HistoryPath, detailResult.Model.History)
		result.Cmd = detailResult.Cmd
		applyDetailResult(&result, detailResult)
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
		result.Model.DetailCursors = msg.detailCursors
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

func applyDetailResult(result *UpdateResult, detailResult detailsview.UpdateResult) {
	if detailResult.Deleted {
		history.DeleteDetailCursors([]string{detailResult.Model.HistoryPath})
		delete(result.Model.DetailCursors, detailResult.Model.HistoryPath)
		nextPath := nextSelectionPathAfterDelete(result.Model.FlatList, detailResult.Model.HistoryPath)
		tree, _ := historytree.Build()
		if nextPath != "" {
			historytree.ExpandAncestorsForPath(tree, nextPath)
		}
		result.Model.Tree = tree
		result.Model.FlatList = historytree.Flatten(tree)
		if nextPath != "" {
			result.Model.Cursor = historytree.FindCursorByPath(result.Model.FlatList, nextPath)
		}
		if result.Model.Cursor >= len(result.Model.FlatList) && len(result.Model.FlatList) > 0 {
			result.Model.Cursor = len(result.Model.FlatList) - 1
		}
		result.Model.Mode = ViewList
		result.Model.Details = detailsview.Model{}
		saveViewState(result.Model.FlatList, result.Model.Cursor)
		result.Cmd = tea.Batch(result.Cmd, historytree.LoadCountsForExpandedNodes(result.Model.Tree))
	} else if detailResult.Quit {
		if result.Model.DetailCursors == nil {
			result.Model.DetailCursors = make(map[string]int)
		}
		result.Model.DetailCursors[detailResult.Model.HistoryPath] = detailResult.Model.Cursor
		history.SaveDetailCursor(detailResult.Model.HistoryPath, detailResult.Model.Cursor)
		result.Model.Mode = ViewList
		result.Model.Details = detailsview.Model{}
		saveViewState(result.Model.FlatList, result.Model.Cursor)
	}
	if detailResult.ScanAllPorts {
		result.ScanAllPorts = true
		result.SelectedHostIP = detailResult.SelectedHostIP
		result.ScanHistoryPath = detailResult.ScanHistoryPath
	}
}
