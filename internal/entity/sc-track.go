package entity

import (
	"encoding/json"
	"fmt"
	"strings"
)

// TrackInfo структура трека
// easyjson:json
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

// SanitizeInfo сохраняет только необходимые данные для отправки на сервер
func (ti *TrackInfo) SanitizeInfo() {
	if strings.Contains(ti.TrackLink, "soundcloud.com") {
		ti.TrackLink = strings.Split(ti.TrackLink, "?")[0]
		ti.TrackTitle = strings.Replace(ti.TrackTitle, "Current track: ", "", 1)
		ti.CoverLink = strings.Replace(ti.CoverLink, "t50x50", "t500x500", 1)
		ti.CoverLink = strings.Replace(ti.CoverLink, "t120x120", "t500x500", 1)
	} else if strings.Contains(ti.TrackLink, "music.youtube.com") {
		ti.CoverLink = strings.Replace(ti.CoverLink, "w60-h60", "w500-h500", 1)
	}
}

func (ti *TrackInfo) PrintIndent() {
	j, _ := json.MarshalIndent(ti, "", "  ")
	fmt.Println(string(j))
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
