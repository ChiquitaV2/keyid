package interfaces

type PlaylistNode struct {
	ID       string
	Name     string
	Children []*PlaylistNode
}

type Client interface {
	LoadPlaylist(name string) Collection
	GetPlaylists() []*PlaylistNode
	GetTrackByTitle(pattern string, from Collection) Item
	GetNowPlaying(collection Collection) Item
	GetCompatibleTracks(track Item, from Collection) Collection
	Suggest(collection Collection) Collection
	Generate(collection Collection) Collection
	Run()
	Close()
}
