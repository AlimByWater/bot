package soundcloud

import (
	"arimadj-helper/internal/entity"
	"arimadj-helper/internal/repository/soundcloud/pkg/mp3"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io"
	"net/http"
)

// Download queries the SoundCloud api and receives a m3u8 file, then binds the segments received into a .mp3 file
func (s *Soundcloud) Download(url string, info entity.TrackInfo) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	// set Non Hacker User Agent
	req.Header.Set("Accept", s.UserAgent)
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cliend do: %w", err)
	}

	// parse html
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("get artwork: %w", err)
	}

	streamURL, err := s.ConstructStreamURL(url, doc)
	if err != nil {
		return "", fmt.Errorf("get artwork: %w", err)
	}

	//songName, err := s.GetTitle(doc)
	//if err != nil {
	//
	//	return fmt.Errorf("get title: %w", err)
	//}

	artwork, err := s.GetArtwork(doc)
	if err != nil {
		return "", fmt.Errorf("get artwork: %w", err)
	}

	// Get the response from the URL
	streamResp, err := http.Get(streamURL)
	if err != nil {
		return "", fmt.Errorf("get streamURL: %w", err)
	}
	defer streamResp.Body.Close()

	// Read the body of the response
	body, err := io.ReadAll(streamResp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	// Unmarshal the JSON into the struct
	var audioResp AudioLink
	err = json.Unmarshal(body, &audioResp)
	if err != nil {
		return "", fmt.Errorf("unmarshal audiolink: %w", err)
	}

	// merge segments
	mp3.Merge(audioResp.URL, info.TrackTitle)

	artworkResp, err := http.Get(artwork)
	image, err := io.ReadAll(artworkResp.Body)
	if err != nil {
		return "", fmt.Errorf("readall artwork: %w", err)
	}

	// set cover image for mp3 file
	mp3.SetTitleArtistCoverImage(info.TrackTitle+".mp3", info.TrackTitle, info.ArtistName, image)
	return info.TrackTitle + ".mp3", nil
}
