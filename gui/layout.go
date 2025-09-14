package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// createWidgets initializes all the UI widgets for the GUI.
func (g *GUI) createWidgets() {
	g.statusBar = widget.NewLabel("Ready")
	g.infoCard = widget.NewCard("Information", "", container.NewVBox(
		widget.NewRichTextFromMarkdown("**No playlist selected**"),
		widget.NewSeparator(),
		widget.NewRichTextFromMarkdown("_Press 'Now Playing' to update_"),
	))
	g.playlistInfoLabel = g.infoCard.Content.(*fyne.Container).Objects[0].(*widget.RichText)
	g.nowPlayingInfoLabel = g.infoCard.Content.(*fyne.Container).Objects[2].(*widget.RichText)

	g.createPlaylistTree()
	g.createActionButtons()
	g.createSuggestionsTable()
	g.createGeneratedTable()
}

// createPlaylistTree creates the playlist tree widget.
func (g *GUI) createPlaylistTree() {
	g.playlistTree = widget.NewTree(
		g.getTreeChildren,
		g.isTreeBranch,
		g.createTreeItem,
		g.updateTreeItem,
	)
	g.playlistTree.OnSelected = g.onPlaylistSelected
	g.playlistTree.Root = ""
}

// createActionButtons creates the main action buttons.
func (g *GUI) createActionButtons() {
	g.refreshBtn = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), g.handleRefresh)
	g.nowPlayingBtn = widget.NewButtonWithIcon("Now Playing", theme.MediaMusicIcon(), g.handleShowNowPlaying)
	g.suggestBtn = widget.NewButtonWithIcon("Get Suggestions", theme.SearchIcon(), g.handleSuggest)
	g.generateBtn = widget.NewButtonWithIcon("Generate Playlist", theme.MediaPlayIcon(), g.handleGenerate)
	g.exportBtn = widget.NewButtonWithIcon("Export M3U", theme.DocumentSaveIcon(), g.handleExport)

	g.suggestBtn.Importance = widget.HighImportance
	g.generateBtn.Importance = widget.HighImportance
}

// createSuggestionsTable creates the table for suggested tracks.
func (g *GUI) createSuggestionsTable() {
	g.suggestionsTable = widget.NewTable(
		func() (int, int) { return len(g.suggestedTracks) + 1, 4 },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			g.updateTrackCell(id, cell, g.suggestedTracks)
		},
	)
	g.suggestionsTable.SetColumnWidth(0, 250) // Title
	g.suggestionsTable.SetColumnWidth(1, 200) // Artist
	g.suggestionsTable.SetColumnWidth(2, 80)  // BPM
	g.suggestionsTable.SetColumnWidth(3, 80)  // Key
}

// createGeneratedTable creates the table for the generated playlist.
func (g *GUI) createGeneratedTable() {
	g.generatedTable = widget.NewTable(
		func() (int, int) { return len(g.generatedTracks) + 1, 4 },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			g.updateTrackCell(id, cell, g.generatedTracks)
		},
	)
	g.generatedTable.SetColumnWidth(0, 250) // Title
	g.generatedTable.SetColumnWidth(1, 200) // Artist
	g.generatedTable.SetColumnWidth(2, 80)  // BPM
	g.generatedTable.SetColumnWidth(3, 80)  // Key
}

// setupLayout assembles the created widgets into the final window layout.
func (g *GUI) setupLayout() {
	// Left Panel
	scrollableTree := container.NewVScroll(g.playlistTree)
	playlistCard := widget.NewCard("Playlists", "", scrollableTree)
	leftPanelBottomButtons := container.NewHBox(g.refreshBtn, g.nowPlayingBtn)
	leftPanel := container.NewBorder(g.infoCard, leftPanelBottomButtons, nil, nil, playlistCard)

	// Right Panel
	buttonBar := container.NewHBox(g.suggestBtn, g.generateBtn, g.exportBtn)
	tabs := container.NewAppTabs(
		container.NewTabItem("Suggestions", g.suggestionsTable),
		container.NewTabItem("Generated Playlist", g.generatedTable),
	)
	tabs.SetTabLocation(container.TabLocationTop)
	rightPanel := container.NewBorder(buttonBar, nil, nil, nil, tabs)

	mainSplit := container.NewHSplit(leftPanel, rightPanel)
	mainSplit.Offset = DefaultSplitOffset
	content := container.NewBorder(nil, g.statusBar, nil, nil, mainSplit)

	g.w.SetContent(content)
}
