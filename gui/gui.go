package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/xdave/keyid/interfaces"
	"go.uber.org/fx"
)

const (
	DefaultWindowWidth  = 1000
	DefaultWindowHeight = 700
	DefaultSplitOffset  = 0.35
	MinColumnWidth      = 120
)

type GUI struct {
	client interfaces.Client
	w      fyne.Window

	// UI Components
	playlistTree      *widget.Tree
	currentTrackCard  *widget.Card
	currentTrackLabel *widget.RichText
	suggestionsCard   *widget.Card
	suggestionsList   *widget.List
	generatedCard     *widget.Card
	generatedList     *widget.List
	statusBar         *widget.Label

	// Action buttons
	suggestBtn  *widget.Button
	generateBtn *widget.Button
	exportBtn   *widget.Button
	refreshBtn  *widget.Button

	// Data
	playlists        []*interfaces.PlaylistNode
	playlistMap      map[string]*interfaces.PlaylistNode
	currentTracks    interfaces.Collection
	suggestedTracks  []interfaces.Item
	generatedTracks  []interfaces.Item
	selectedPlaylist *interfaces.PlaylistNode
}

func Show(client interfaces.Client, shutdowner fx.Shutdowner) {
	a := app.NewWithID("com.github.xdave.keyid")
	a.Settings().SetTheme(theme.DarkTheme())

	w := a.NewWindow("KeyID - DJ Track Suggestion & Playlist Generator")
	w.SetOnClosed(func() {
		log.Println("Application shutting down...")
		shutdowner.Shutdown()
	})

	gui := &GUI{
		client:      client,
		w:           w,
		playlistMap: make(map[string]*interfaces.PlaylistNode),
	}

	if err := gui.initialize(); err != nil {
		log.Printf("Failed to initialize GUI: %v", err)
		dialog.ShowError(fmt.Errorf("initialization failed: %v", err), w)
		return
	}

	gui.setupUI()
	gui.setInitialState()

	w.Resize(fyne.NewSize(DefaultWindowWidth, DefaultWindowHeight))
	w.CenterOnScreen()
	w.ShowAndRun()
}

func (g *GUI) initialize() error {
	g.updateStatus("Loading playlists...")

	playlists := g.client.GetPlaylists()
	if len(playlists) == 0 {
		log.Println("Warning: No playlists found from client, using test data")
		// Fallback test data for development
		playlists = []*interfaces.PlaylistNode{
			{
				ID:   "test1",
				Name: "House Music",
				Children: []*interfaces.PlaylistNode{
					{ID: "house1", Name: "Deep House"},
					{ID: "house2", Name: "Tech House"},
				},
			},
			{
				ID:   "test2",
				Name: "Electronic",
				Children: []*interfaces.PlaylistNode{
					{ID: "elec1", Name: "Synthwave"},
					{ID: "elec2", Name: "Ambient"},
				},
			},
		}
	}

	g.playlists = playlists
	g.buildPlaylistMap(g.playlists)

	log.Printf("Successfully loaded %d root playlists", len(g.playlists))
	return nil
}

func (g *GUI) buildPlaylistMap(nodes []*interfaces.PlaylistNode) {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		g.playlistMap[node.ID] = node
		if len(node.Children) > 0 {
			g.buildPlaylistMap(node.Children)
		}
	}
}

func (g *GUI) setupUI() {
	g.createWidgets()
	g.setupLayout()
}

func (g *GUI) createWidgets() {
	g.createStatusBar()
	g.createPlaylistTree()
	g.createCurrentTrackCard()
	g.createActionButtons()
	g.createSuggestionsCard()
	g.createGeneratedCard()
}

func (g *GUI) createStatusBar() {
	g.statusBar = widget.NewLabel("Ready")
	g.statusBar.Alignment = fyne.TextAlignLeading
}

func (g *GUI) createPlaylistTree() {
	g.playlistTree = widget.NewTree(
		g.getTreeChildren,
		//func(id widget.TreeNodeID) []widget.TreeNodeID {
		//	println("called")
		//	return {widget.TreeNodeID({"123"})
		//},
		g.isTreeBranch,
		g.createTreeItem,
		g.updateTreeItem,
	)

	g.playlistTree.OnSelected = g.onPlaylistSelected
	g.playlistTree.Root = ""
	g.playlistTree.ExtendBaseWidget(g.playlistTree)
	g.playlistTree.Refresh()
}

func (g *GUI) createCurrentTrackCard() {
	g.currentTrackLabel = widget.NewRichTextFromMarkdown("**No track selected**")
	g.currentTrackCard = widget.NewCard("Current Track", "",
		container.NewVBox(g.currentTrackLabel))
}

func (g *GUI) createActionButtons() {
	g.refreshBtn = widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), g.handleRefresh)
	g.suggestBtn = widget.NewButtonWithIcon("Get Suggestions", theme.SearchIcon(), g.handleSuggest)
	g.generateBtn = widget.NewButtonWithIcon("Generate Playlist", theme.MediaPlayIcon(), g.handleGenerate)
	g.exportBtn = widget.NewButtonWithIcon("Export M3U", theme.DocumentSaveIcon(), g.handleExport)

	// Set button importance
	g.suggestBtn.Importance = widget.HighImportance
	g.generateBtn.Importance = widget.HighImportance
}

func (g *GUI) createSuggestionsCard() {
	g.suggestionsList = widget.NewList(
		func() int { return len(g.suggestedTracks) },
		g.createTrackItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			g.updateTrackItem(i, o, g.suggestedTracks)
		},
	)

	g.suggestionsCard = widget.NewCard("Track Suggestions", "",
		container.NewBorder(nil, nil, nil, nil, g.suggestionsList))
}

func (g *GUI) createGeneratedCard() {
	g.generatedList = widget.NewList(
		func() int { return len(g.generatedTracks) },
		g.createTrackItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			g.updateTrackItem(i, o, g.generatedTracks)
		},
	)

	g.generatedCard = widget.NewCard("Generated Playlist", "",
		container.NewBorder(nil, nil, nil, nil, g.generatedList))
}

func (g *GUI) setupLayout() {

	scrollableTree := container.NewVScroll(g.playlistTree)

	playlistCard := widget.NewCard("Playlists", "", scrollableTree)

	leftPanel := container.NewBorder(
		g.currentTrackCard,
		container.NewHBox(g.refreshBtn),
		nil,
		nil,
		playlistCard,
	)

	buttonBar := container.NewHBox(
		g.suggestBtn,
		g.generateBtn,
		g.exportBtn,
	)

	tabs := container.NewAppTabs(
		container.NewTabItem("Suggestions", g.suggestionsList),
		container.NewTabItem("Generated Playlist", g.generatedList),
	)

	rightPanel := container.NewBorder(
		buttonBar, nil, nil, nil,
		tabs,
	)

	mainSplit := container.NewHSplit(leftPanel, rightPanel)
	mainSplit.Offset = DefaultSplitOffset

	content := container.NewBorder(
		nil, g.statusBar, nil, nil,
		mainSplit,
	)

	g.w.SetContent(content)
}

func (g *GUI) setInitialState() {
	g.updateButtonStates()
	g.updateStatus("Select a playlist to begin")
}

func (g *GUI) updateButtonStates() {
	hasPlaylist := g.selectedPlaylist != nil && g.currentTracks != nil
	//hasTracks := len(g.suggestedTracks) > 0 || len(g.generatedTracks) > 0

	g.suggestBtn.Enable()
	g.generateBtn.Enable()

	if !hasPlaylist {
		g.suggestBtn.Disable()
		g.generateBtn.Disable()
	}

	g.exportBtn.Enable()
	if len(g.generatedTracks) == 0 {
		g.exportBtn.Disable()
	}
}

// Tree widget methods
func (g *GUI) getTreeChildren(id widget.TreeNodeID) []widget.TreeNodeID {
	nodeID := string(id) // Convert TreeNodeID to string for map keys/comparisons

	if nodeID == "" {
		// Return root level playlists
		children := make([]widget.TreeNodeID, 0, len(g.playlists))
		for _, playlist := range g.playlists {
			if playlist != nil {
				// Convert string ID back to TreeNodeID for the return slice
				children = append(children, widget.TreeNodeID(playlist.ID))
			}
		}
		return children
	}

	node, exists := g.playlistMap[nodeID]
	if !exists || node == nil {
		return []widget.TreeNodeID{}
	}

	children := make([]widget.TreeNodeID, 0, len(node.Children))
	for _, child := range node.Children {
		if child != nil {
			// Convert string ID back to TreeNodeID for the return slice
			children = append(children, widget.TreeNodeID(child.ID))
		}
	}
	return children
}

func (g *GUI) isTreeBranch(id widget.TreeNodeID) bool {
	if id == "" {
		return true
	}

	nodeID := string(id) // Convert TreeNodeID to string for map lookup
	node, exists := g.playlistMap[nodeID]
	return exists && node != nil && len(node.Children) > 0
}

func (g *GUI) createTreeItem(branch bool) fyne.CanvasObject {
	icon := theme.DocumentIcon()
	if branch {
		icon = theme.FolderIcon()
	}
	return container.NewHBox(
		widget.NewIcon(icon),
		widget.NewLabel("Loading..."),
	)
}

func (g *GUI) updateTreeItem(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
	container, ok := obj.(*fyne.Container)
	if !ok || len(container.Objects) < 2 {
		return
	}

	icon, ok := container.Objects[0].(*widget.Icon)
	if !ok {
		return
	}

	label, ok := container.Objects[1].(*widget.Label)
	if !ok {
		return
	}

	node, exists := g.playlistMap[id]
	if !exists || node == nil {
		label.SetText("Unknown")
		return
	}

	// Update icon based on branch status
	if branch {
		icon.SetResource(theme.FolderIcon())
	} else {
		icon.SetResource(theme.MediaMusicIcon())
	}

	label.SetText(node.Name)
}

// Event handlers
func (g *GUI) onPlaylistSelected(id widget.TreeNodeID) {
	if id == "" {
		return
	}

	nodeID := string(id)
	node, exists := g.playlistMap[nodeID]
	if !exists || node == nil {
		g.showError("Invalid playlist selection")
		return
	}

	// Don't load folder nodes
	if len(node.Children) > 0 {
		g.updateStatus(fmt.Sprintf("Folder selected: %s", node.Name))
		return
	}

	g.selectedPlaylist = node
	g.loadPlaylist(node)
}

func (g *GUI) loadPlaylist(node *interfaces.PlaylistNode) {
	g.updateStatus(fmt.Sprintf("Loading playlist: %s...", node.Name))
	g.currentTrackLabel.ParseMarkdown("**Loading...**")

	// Clear previous data
	g.suggestedTracks = []interfaces.Item{}
	g.generatedTracks = []interfaces.Item{}
	g.suggestionsList.Refresh()
	g.generatedList.Refresh()

	// Load tracks
	tracks := g.client.LoadPlaylist(node.Name)
	if tracks == nil {
		g.showError(fmt.Sprintf("Failed to load playlist: %s", node.Name))
		g.currentTrackLabel.ParseMarkdown("**Failed to load playlist**")
		g.updateStatus("Ready")
		g.updateButtonStates()
		return
	}

	g.currentTracks = tracks
	trackCount := tracks.Len()

	g.currentTrackLabel.ParseMarkdown(fmt.Sprintf("**Playlist:** %s  \n**Tracks:** %d",
		node.Name, trackCount))

	g.updateStatus(fmt.Sprintf("Loaded %d tracks from %s", trackCount, node.Name))
	g.updateButtonStates()

	log.Printf("Successfully loaded playlist '%s' with %d tracks", node.Name, trackCount)
}

func (g *GUI) handleRefresh() {
	g.updateStatus("Refreshing...")

	if err := g.initialize(); err != nil {
		g.showError(fmt.Sprintf("Failed to refresh: %v", err))
		return
	}

	g.playlistTree.Refresh()
	g.updateCurrentTrack()
	g.updateStatus("Playlists refreshed")
}

func (g *GUI) handleSuggest() {
	if g.currentTracks == nil {
		g.showError("Please select a playlist first")
		return
	}

	g.updateStatus("Getting track suggestions...")

	suggestedCollection := g.client.Suggest(g.currentTracks)
	if suggestedCollection == nil {
		g.suggestedTracks = []interfaces.Item{}
		g.updateStatus("No suggestions found")
	} else {
		g.suggestedTracks = suggestedCollection.Items()
		g.updateStatus(fmt.Sprintf("Found %d suggested tracks", len(g.suggestedTracks)))
	}

	g.suggestionsList.Refresh()
	g.updateCurrentTrack()
}

func (g *GUI) handleGenerate() {
	if g.currentTracks == nil {
		g.showError("Please select a playlist first")
		return
	}

	g.updateStatus("Generating playlist...")

	generatedCollection := g.client.Generate(g.currentTracks)
	if generatedCollection == nil {
		g.generatedTracks = []interfaces.Item{}
		g.updateStatus("Failed to generate playlist")
		g.showError("Failed to generate playlist")
	} else {
		g.generatedTracks = generatedCollection.Items()
		g.updateStatus(fmt.Sprintf("Generated playlist with %d tracks", len(g.generatedTracks)))
	}

	g.generatedList.Refresh()
	g.updateButtonStates()
}

func (g *GUI) handleExport() {
	if len(g.generatedTracks) == 0 {
		dialog.ShowInformation("Nothing to Export",
			"Please generate a playlist first.", g.w)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			g.showError(fmt.Sprintf("File save error: %v", err))
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		if err := g.writeM3UFile(writer); err != nil {
			g.showError(fmt.Sprintf("Export failed: %v", err))
			return
		}

		dialog.ShowInformation("Export Successful",
			"Playlist exported successfully!", g.w)
		g.updateStatus("Playlist exported successfully")
	}, g.w)
}

// Helper methods
func (g *GUI) updateCurrentTrack() {
	if g.currentTracks == nil {
		return
	}

	currentTrack := g.client.GetNowPlaying(g.currentTracks)
	if currentTrack != nil {
		trackInfo := fmt.Sprintf("**Current Track:** %s  \n**Artist:** %s  \n**BPM:** %.1f  \n**Key:** %s",
			currentTrack.GetTitle(),
			currentTrack.GetArtist(),
			currentTrack.GetBPM(),
			currentTrack.GetScale().String())
		g.currentTrackLabel.ParseMarkdown(trackInfo)
	}
}

func (g *GUI) createTrackItem() fyne.CanvasObject {
	titleLabel := widget.NewLabel("Title")
	titleLabel.Truncation = fyne.TextTruncateEllipsis

	artistLabel := widget.NewLabel("Artist")
	artistLabel.Truncation = fyne.TextTruncateEllipsis

	bpmLabel := widget.NewLabel("BPM")
	bpmLabel.Alignment = fyne.TextAlignCenter

	keyLabel := widget.NewLabel("Key")
	keyLabel.Alignment = fyne.TextAlignCenter

	// Create a grid layout for better organization
	return container.NewHBox(
		container.NewWithoutLayout(titleLabel),
		container.NewWithoutLayout(artistLabel),
		container.NewWithoutLayout(bpmLabel),
		container.NewWithoutLayout(keyLabel),
	)
}

func (g *GUI) updateTrackItem(i widget.ListItemID, obj fyne.CanvasObject, tracks []interfaces.Item) {
	if i >= len(tracks) {
		return
	}

	track := tracks[i]
	container, ok := obj.(*fyne.Container)
	if !ok || len(container.Objects) < 4 {
		return
	}

	// Update each label
	if titleContainer, ok := container.Objects[0].(*fyne.Container); ok && len(titleContainer.Objects) > 0 {
		if titleLabel, ok := titleContainer.Objects[0].(*widget.Label); ok {
			title := track.GetTitle()
			if len(title) > 30 {
				title = title[:27] + "..."
			}
			titleLabel.SetText(title)
		}
	}

	if artistContainer, ok := container.Objects[1].(*fyne.Container); ok && len(artistContainer.Objects) > 0 {
		if artistLabel, ok := artistContainer.Objects[0].(*widget.Label); ok {
			artist := track.GetArtist()
			if len(artist) > 25 {
				artist = artist[:22] + "..."
			}
			artistLabel.SetText(artist)
		}
	}

	if bpmContainer, ok := container.Objects[2].(*fyne.Container); ok && len(bpmContainer.Objects) > 0 {
		if bpmLabel, ok := bpmContainer.Objects[0].(*widget.Label); ok {
			bpmLabel.SetText(fmt.Sprintf("%.1f", track.GetBPM()))
		}
	}

	if keyContainer, ok := container.Objects[3].(*fyne.Container); ok && len(keyContainer.Objects) > 0 {
		if keyLabel, ok := keyContainer.Objects[0].(*widget.Label); ok {
			keyLabel.SetText(track.GetScale().String())
		}
	}
}

func (g *GUI) writeM3UFile(writer fyne.URIWriteCloser) error {
	if _, err := fmt.Fprintln(writer, "#EXTM3U"); err != nil {
		return err
	}

	for _, track := range g.generatedTracks {
		if track == nil {
			continue
		}

		// Write track info
		if _, err := fmt.Fprintf(writer, "#EXTINF:-1,%s - %s\n",
			track.GetArtist(), track.GetTitle()); err != nil {
			return err
		}

		// Write track path
		if _, err := fmt.Fprintln(writer, track.GetPath()); err != nil {
			return err
		}
	}

	return nil
}

func (g *GUI) updateStatus(message string) {
	if g.statusBar != nil {
		g.statusBar.SetText(message)
	}
	log.Printf("Status: %s", message)
}

func (g *GUI) showError(message string) {
	log.Printf("Error: %s", message)
	g.updateStatus(fmt.Sprintf("Error: %s", message))
	dialog.ShowError(fmt.Errorf(message), g.w)
}
