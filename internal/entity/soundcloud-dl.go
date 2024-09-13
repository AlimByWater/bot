package entity

var ResolveApiUrl = "https://api-widget.soundcloud.com/resolve?"

// easyjson:json
type Transcode struct {
	ApiUrl  string `json:"url"`
	Quality string `json:"quality"`
	Format  Format `json:"format"`
}

// easyjson:json
type Format struct {
	Protocol string `json:"protocol"`
	MimeType string `json:"mime_type"`
}

// easyjson:json
type SoundData struct {
	Id           int64      `json:"id"`
	Title        string     `json:"title"`
	CreatedAt    string     `json:"created_at"`
	Duration     int64      `json:"duration"`
	Kind         string     `json:"kind"`
	PermalinkUrl string     `json:"permalink_url"`
	UserId       int64      `json:"user_id"`
	ArtworkUrl   string     `json:"artwork_url"`
	Genre        string     `json:"genre"`
	Transcodes   Transcodes `json:"media"`
	LikesCount   int        `json:"likes_count"`
	Downloadable bool       `json:"downloadable"`
	Description  string     `json:"description,omitempty"`
}

// easyjson:json
type Transcodes struct {
	Transcodings []Transcode `json:"transcodings"`
}

// easyjson:json
type Media struct {
	Url string `json:"url"`
}

// easyjson:json
type DownloadTrack struct {
	Url       string
	Size      int
	Data      []byte
	Quality   string
	Ext       string
	SoundData *SoundData
}

// easyjson:json
type SearchResult struct {
	Sounds []SoundData `json:"collection"`
	Next   string      `json:"next_href"`
}
