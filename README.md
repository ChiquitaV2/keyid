# keyid

Uses the Camelot system to either:

- suggest the next track (based on your current Master deck)
  - Rekordbox takes 60 seconds (by default) to add the Master deck track to your history. See [here](https://assets-global.website-files.com/624b7d52289aa4ed9d117a25/6395995aaa7b424b012bfea7_Screen%20Shot%202022-12-11%20at%2012.47.54%20AM.png) for how to change this setting
- generate a new playlist from the given playlist name (or your whole collection) with the selections in a compatible key and tempo.

This assumes you have analyzed your collection using the Alphanumeric key notation.

# Why

- Q: Rekordbox already has a "Traffic Light" system that can suggest tracks to play that are in key, why do this?
- A: It sucks, and this program has a better set of rules that finds stuff that works more harmonically.

# Prerequisites

- Install [rekordbox](https://rekordbox.com/en/)
- Set the key display format to [Alphanumeric](https://support.pioneerdj.com/hc/en-us/articles/8943219092761-Can-I-change-the-display-format-for-keys)
- Analyze your tracks. I recommend using [Mixed In Key](https://mixedinkey.com/integration/rekordbox-integration/) (other key detection software makes this program mostly useless because they're almost always wrong)

# Build instructions for macOS (only tested on Sonoma & Sequoia)

## Dependencies

- Install the [command-line tools](https://mac.install.guide/commandlinetools/4)
- Install [homebrew](https://brew.sh/)
- Install openssl: `brew install openssl@3`
  - Then, create symbolic links to the openssl library in `/usr/local`:
    - `sudo ln -svf /opt/homebrew/include/openssl /usr/local/include/`
    - `sudo ln -svf /opt/homebrew/lib/libcrypto.* /usr/local/lib/`
- Install [go 1.22.x](https://go.dev/doc/install)

## Build

- clone this repo: `git clone https://github.com/xdave/keyid.git`
- switch to it's directory: `cd keyid`
- install build dependencies: `go get`
- build the app: `go build .`
- run the app: `./keyid --help` (with `--help` to get usage instructions)

# Build instructions for Windows

## Dependencies

- [TODO]

## Build

- [TODO]

# Linux

- At the moment, only the platforms upon which Rekordbox is supported work (Windows, Mac)

# Usage

```
Usage of ./keyid:
  -debug
        Enable debug logging
  -mode string
        One of 'suggest' or 'generate' (default "suggest")
  -playlist string
        Name of Rekordbox Playlist to use (uses whole collection by default)
  -random
        Randomize playlist before 'generate'
  -startWith string
        Some part of the Track Title to start with in 'generate' mode (otherwise
        starts with first track in provided 'playlist')
```

## Examples

- To suggest the next track based on what you're currently playing (lists all compatible tracks from your collection):

```
./keyid
```

- To use an existing playlist for the pool of tracks to find:

```
./keyid -playlist 'My Cool Playlist 2024'
```

- You can also suggest based on the name of a track you want, instead of the one that's playing

```
./keyid -playlist 'My Cool Playlist 2024' -startWith 'Cafe Del Mar
```

- To generate a new playlist based on your whole collection (also accepts `-playlist`):

```
./keyid -mode generate -playlist 'My Cool Playlist 2024' -startWith 'Cafe Del Mar'
```

- NOTE: You can provide a track to start with from your source playlist when in `generate` mode.
- NOTE: Generate mode can randomize the order of the tracks it looks at in the provided playlist, so you can run it multiple times to get a new selection if it doesn't generate something useful (see `-random` flag)
- NOTE: Track printout has 4 columns, BPM, Key, Energy, and Artist+Title, for example:
  - `122 10A     6       Serious Dancers - In The Beginning (Hernan Cattaneo & Simply City Remix)`

# Do you _really_ use this?

- Yes, [I do](https://soundcloud.com/davidgradwell/sets/gradwell-radio). I used to do this manually in my head.

# Roadmap

- [x] Generate .m3u playlists in `generate` mode
- [ ] Multiplatform downloadable builds
- [ ] Support for other DJ Software (Serato, Virtual DJ, Engine DJ, Mixxx, etc)
- [ ] Better documentation
