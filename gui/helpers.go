package gui

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/xdave/keyid/interfaces"
)

// buildPlaylistMap recursively creates a flat map for quick lookup of playlists by ID.
func (g *GUI) buildPlaylistMap(nodes []*interfaces.PlaylistNode) {
	for _, node := range nodes {
		if node != nil {
			g.playlistMap[node.ID] = node
			if len(node.Children) > 0 {
				g.buildPlaylistMap(node.Children)
			}
		}
	}
}

// writeM3UFile writes the generated tracks to a file in M3U format.
func (g *GUI) writeM3UFile(writer fyne.URIWriteCloser) error {
	if _, err := fmt.Fprintln(writer, "#EXTM3U"); err != nil {
		return err
	}
	for _, track := range g.generatedTracks {
		if track != nil {
			if _, err := fmt.Fprintf(writer, "#EXTINF:-1,%s - %s\n", track.GetArtist(), track.GetTitle()); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(writer, track.GetPath()); err != nil {
				return err
			}
		}
	}
	return nil
}

// updateStatus updates the text in the status bar and logs the message.
func (g *GUI) updateStatus(message string) {
	if g.statusBar != nil {
		g.statusBar.SetText(message)
	}
	log.Printf("Status: %s", message)
}

// showError displays an error dialog to the user and logs the error.
func (g *GUI) showError(message string) {
	log.Printf("Error: %s", message)
	g.updateStatus(fmt.Sprintf("Error: %s", message))
	dialog.ShowError(fmt.Errorf(message), g.w)
}
