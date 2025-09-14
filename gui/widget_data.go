package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/xdave/keyid/interfaces"
)

// Tree data functions

func (g *GUI) getTreeChildren(id widget.TreeNodeID) []widget.TreeNodeID {
	if id == "" {
		children := make([]widget.TreeNodeID, 0, len(g.playlists))
		for _, playlist := range g.playlists {
			children = append(children, playlist.ID)
		}
		return children
	}
	node, exists := g.playlistMap[id]
	if !exists || node == nil {
		return []widget.TreeNodeID{}
	}
	children := make([]widget.TreeNodeID, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child.ID)
	}
	return children
}

func (g *GUI) isTreeBranch(id widget.TreeNodeID) bool {
	if id == "" {
		return true // The root is always a branch
	}
	node, exists := g.playlistMap[id]
	return exists && node != nil && len(node.Children) > 0
}

func (g *GUI) createTreeItem(isBranch bool) fyne.CanvasObject {
	icon := theme.DocumentIcon()
	if isBranch {
		icon = theme.FolderIcon()
	}
	return container.NewHBox(widget.NewIcon(icon), widget.NewLabel("Loading..."))
}

func (g *GUI) updateTreeItem(id widget.TreeNodeID, isBranch bool, item fyne.CanvasObject) {
	node, exists := g.playlistMap[id]
	if !exists {
		return
	}
	hbox := item.(*fyne.Container)
	label := hbox.Objects[1].(*widget.Label)
	label.SetText(node.Name)

	icon := hbox.Objects[0].(*widget.Icon)
	if isBranch {
		icon.SetResource(theme.FolderOpenIcon())
	} else {
		icon.SetResource(theme.MediaMusicIcon())
	}
}

func (g *GUI) updateTrackCell(id widget.TableCellID, cell fyne.CanvasObject, tracks []interfaces.Item) {
	label := cell.(*widget.Label)
	if id.Row == 0 { // Header row
		label.TextStyle.Bold = true
		switch id.Col {
		case 0:
			label.SetText("Title")
		case 1:
			label.SetText("Artist")
		case 2:
			label.SetText("BPM")
		case 3:
			label.SetText("Key")
		}
		return
	}

	track := tracks[id.Row-1]
	label.TextStyle.Bold = false
	switch id.Col {
	case 0:
		label.SetText(track.GetTitle())
	case 1:
		label.SetText(track.GetArtist())
	case 2:
		label.SetText(fmt.Sprintf("%.1f", track.GetBPM()))
	case 3:
		label.SetText(track.GetScale().String())
	}
}
