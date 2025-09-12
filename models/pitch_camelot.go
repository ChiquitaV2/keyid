package models

// PitchToCamelot maps key to Camelot notation
var PitchToCamelot = map[string]string{
	// Major keys (B)
	"C":  "8B",
	"C#": "1B",
	"Db": "1B",
	"D":  "3B",
	"D#": "4B",
	"Eb": "4B",
	"E":  "5B",
	"F":  "6B",
	"F#": "7B",
	"Gb": "7B",
	"G":  "8B",
	"G#": "9B",
	"Ab": "9B",
	"A":  "10B",
	"A#": "11B",
	"Bb": "11B",
	"B":  "12B",

	// Minor keys (A)
	"Am":  "8A",
	"A#m": "1A",
	"Bbm": "1A",
	"Bm":  "3A",
	"Cm":  "4A",
	"C#m": "5A",
	"Dbm": "5A",
	"Dm":  "6A",
	"D#m": "7A",
	"Ebm": "7A",
	"Em":  "8A",
	"Fm":  "9A",
	"F#m": "10A",
	"Gbm": "10A",
	"Gm":  "11A",
	"G#m": "12A",
	"Abm": "12A",
}
