package entity

import "strings"

// TrackInfo структура трека
type TrackInfo struct {
	TrackTitle string `json:"trackTitle"`
	ArtistName string `json:"artistName"`
	TrackLink  string `json:"trackLink"`
	Duration   string `json:"duration"`
	Artwork    []byte `json:"artwork"`
	ArtworkUrl string `json:"artworkUrl"`
}

func (t *TrackInfo) Format() {
	t.TrackTitle = strings.Replace(t.TrackTitle, "Current track: ", "", 1)

	t.TrackTitle = formatEscapeChars(t.TrackTitle)
	t.ArtistName = formatEscapeChars(t.ArtistName)
	t.TrackLink = formatEscapeChars(t.TrackLink)
	t.Duration = formatEscapeChars(t.Duration)
}

// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
func formatEscapeChars(oldS string) string {
	s := oldS
	s = strings.ReplaceAll(s, `_`, `\_`)
	s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, `[`, `\[`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	s = strings.ReplaceAll(s, `~`, `\~`)
	//s = strings.ReplaceAll(s, "`", "\`")
	s = strings.ReplaceAll(s, `>`, `\>`)
	s = strings.ReplaceAll(s, `#`, `\#`)
	s = strings.ReplaceAll(s, `+`, `\+`)
	s = strings.ReplaceAll(s, `-`, `\-`)
	s = strings.ReplaceAll(s, `=`, `\=`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	s = strings.ReplaceAll(s, `{`, `\{`)
	s = strings.ReplaceAll(s, `}`, `\}`)
	s = strings.ReplaceAll(s, `.`, `\.`)
	s = strings.ReplaceAll(s, `!`, `\!`)

	return s
}
