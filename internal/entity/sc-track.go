package entity

import "strings"

// TrackInfo структура трека
type TrackInfo struct {
	TrackTitle string   `json:"trackTitle"`
	ArtistName string   `json:"artistName"`
	TrackLink  string   `json:"trackLink"`
	Duration   string   `json:"duration"`
	CoverLink  string   `json:"artworkUrl"`
	Tags       []string `json:"tags"`
}

func (ti *TrackInfo) Format() TrackInfo {
	t := TrackInfo{}

	t.TrackTitle = strings.Replace(ti.TrackTitle, "Current track: ", "", 1)

	t.TrackTitle = formatEscapeChars(ti.TrackTitle)
	t.ArtistName = formatEscapeChars(ti.ArtistName)
	t.TrackLink = formatEscapeChars(ti.TrackLink)
	t.Duration = formatEscapeChars(ti.Duration)

	t.CoverLink = formatEscapeChars(ti.CoverLink)
	for _, tag := range ti.Tags {
		t.Tags = append(t.Tags, formatEscapeChars(tag))
	}
	return t
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
