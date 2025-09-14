package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/xdave/keyid/interfaces"
)

// onPlaylistSelected handles the event when a user selects a playlist from the tree.
func (g *GUI) onPlaylistSelected(id widget.TreeNodeID) {
	if id == "" {
		return
	}
	node, exists := g.playlistMap[string(id)]
	if !exists || node == nil {
		g.showError("Invalid playlist selection")
		return
	}
	if len(node.Children) > 0 {
		g.updateStatus(fmt.Sprintf("Folder selected: %s", node.Name))
		return
	}
	g.selectedPlaylist = node
	g.loadPlaylist(node)
}

// loadPlaylist fetches track data for the selected playlist and updates the UI.
func (g *GUI) loadPlaylist(node *interfaces.PlaylistNode) {
	g.updateStatus(fmt.Sprintf("Loading playlist: %s...", node.Name))
	g.playlistInfoLabel.ParseMarkdown("**Loading...**")
	g.nowPlayingInfoLabel.ParseMarkdown("_Press 'Now Playing' to update_")

	g.suggestedTracks = []interfaces.Item{}
	g.generatedTracks = []interfaces.Item{}
	g.suggestionsTable.Refresh()
	g.generatedTable.Refresh()

	tracks := g.client.LoadPlaylist(node.Name)
	if tracks == nil {
		g.showError(fmt.Sprintf("Failed to load playlist: %s", node.Name))
		g.playlistInfoLabel.ParseMarkdown("**Failed to load playlist**")
		g.currentTracks = nil
	} else {
		g.currentTracks = tracks
		trackCount := tracks.Len()
		g.playlistInfoLabel.ParseMarkdown(fmt.Sprintf("**Playlist:** %s  \n**Tracks:** %d", node.Name, trackCount))
		g.updateStatus(fmt.Sprintf("Loaded %d tracks from %s", trackCount, node.Name))
		log.Printf("Successfully loaded playlist '%s' with %d tracks", node.Name, trackCount)
	}
	g.updateButtonStates()
}

// handleRefresh reloads all playlist data from the client.
func (g *GUI) handleRefresh() {
	g.updateStatus("Refreshing...")
	if err := g.initialize(); err != nil {
		g.showError(fmt.Sprintf("Failed to refresh: %v", err))
		return
	}
	g.playlistTree.Refresh()
	g.nowPlayingInfoLabel.ParseMarkdown("_Press 'Now Playing' to update_")
	g.updateStatus("Refreshed playlists")
}

// handleShowNowPlaying gets and displays the currently playing track asynchronously.
func (g *GUI) handleShowNowPlaying() {
	// Immediately update the UI to give feedback that work is starting.
	g.updateStatus("Getting current track...")
	g.nowPlayingInfoLabel.ParseMarkdown("_Loading now playing..._")

	// Run the potentially long-running operation in a separate goroutine.
	go func() {
		currentTrack := g.client.GetNowPlaying(g.currentTracks)

		// Once the data is retrieved, update the UI elements.
		// Fyne's widget operations are thread-safe.
		if currentTrack == nil {
			g.nowPlayingInfoLabel.ParseMarkdown("**No track is currently playing**")
			g.updateStatus("No track playing")
			return
		}

		trackInfo := fmt.Sprintf("**Title:** %s  \n**Artist:** %s  \n**BPM:** %.1f  \n**Key:** %s",
			currentTrack.GetTitle(), currentTrack.GetArtist(), currentTrack.GetBPM(), currentTrack.GetScale().String())

		g.nowPlayingInfoLabel.ParseMarkdown(trackInfo)
		g.updateStatus(fmt.Sprintf("Now Playing: %s", currentTrack.GetTitle()))
	}()
}

// handleSuggest gets track suggestions based on the current playlist.
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
	g.suggestionsTable.Refresh()
}

// handleGenerate creates a new playlist from the suggested tracks.
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
	g.generatedTable.Refresh()
	g.updateButtonStates()
}

// handleExport saves the generated playlist to an M3U file.
func (g *GUI) handleExport() {
	if len(g.generatedTracks) == 0 {
		dialog.ShowInformation("Nothing to Export", "Please generate a playlist first.", g.w)
		return
	}
	fileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			g.showError(fmt.Sprintf("File save error: %v", err))
			return
		}
		if writer == nil { // User cancelled
			return
		}
		defer writer.Close()

		if err := g.writeM3UFile(writer); err != nil {
			g.showError(fmt.Sprintf("Export failed: %v", err))
			return
		}
		dialog.ShowInformation("Export Successful", "Playlist exported successfully!", g.w)
		g.updateStatus("Playlist exported successfully")
	}, g.w)

	fileSaveDialog.SetFileName("generated_playlist.m3u")
	fileSaveDialog.Show()
}
