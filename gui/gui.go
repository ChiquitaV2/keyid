package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/xdave/keyid/interfaces"
	"go.uber.org/fx"
)

// GUI holds the application's state and UI components.
type GUI struct {
	client interfaces.Client
	w      fyne.Window

	// UI Components
	playlistTree        *widget.Tree
	infoCard            *widget.Card
	playlistInfoLabel   *widget.RichText
	nowPlayingInfoLabel *widget.RichText
	suggestionsTable    *widget.Table
	generatedTable      *widget.Table
	statusBar           *widget.Label

	// Action buttons
	suggestBtn    *widget.Button
	generateBtn   *widget.Button
	exportBtn     *widget.Button
	refreshBtn    *widget.Button
	nowPlayingBtn *widget.Button

	// Data
	playlists        []*interfaces.PlaylistNode
	playlistMap      map[string]*interfaces.PlaylistNode
	currentTracks    interfaces.Collection
	suggestedTracks  []interfaces.Item
	generatedTracks  []interfaces.Item
	selectedPlaylist *interfaces.PlaylistNode
}

// Show initializes and runs the GUI application.
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

// initialize loads the initial data for the application.
func (g *GUI) initialize() error {
	g.updateStatus("Loading playlists...")

	clientPlaylists := g.client.GetPlaylists()
	playlists := make([]*interfaces.PlaylistNode, 0)
	for _, p := range clientPlaylists {
		if p != nil {
			playlists = append(playlists, p)
		}
	}

	// Fallback for testing if no playlists are found
	if len(playlists) == 0 {
		log.Println("DEBUG: No valid playlists found from client. Activating fallback test data.")
		playlists = []*interfaces.PlaylistNode{
			{ID: "test1", Name: "House Music", Children: []*interfaces.PlaylistNode{{ID: "house1", Name: "Deep House"}, {ID: "house2", Name: "Tech House"}}},
			{ID: "test2", Name: "Electronic", Children: []*interfaces.PlaylistNode{{ID: "elec1", Name: "Synthwave"}, {ID: "elec2", Name: "Ambient"}}},
		}
	}

	g.playlists = playlists
	g.buildPlaylistMap(g.playlists)

	log.Printf("Successfully loaded %d root playlists into g.playlists", len(g.playlists))
	return nil
}

// setupUI creates and lays out all the UI components.
func (g *GUI) setupUI() {
	g.createWidgets()
	g.setupLayout()
	g.playlistTree.Refresh()
}

// setInitialState sets the UI to its default state after startup.
func (g *GUI) setInitialState() {
	g.updateButtonStates()
	g.updateStatus("Select a playlist to begin")
}

// updateButtonStates enables or disables buttons based on the current app state.
func (g *GUI) updateButtonStates() {
	hasPlaylist := g.selectedPlaylist != nil && g.currentTracks != nil

	g.suggestBtn.Enable()
	g.generateBtn.Enable()
	g.nowPlayingBtn.Enable()

	if !hasPlaylist {
		g.suggestBtn.Disable()
		g.generateBtn.Disable()
		g.nowPlayingBtn.Disable()
	}

	g.exportBtn.Enable()
	if len(g.generatedTracks) == 0 {
		g.exportBtn.Disable()
	}
}
