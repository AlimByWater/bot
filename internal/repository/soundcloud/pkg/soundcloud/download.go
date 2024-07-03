package soundcloud

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"arimadj-helper/internal/repository/soundcloud/pkg/mp3"
	"context"
	"encoding/json"
	"fmt"
	"github.com/antchfx/htmlquery"
	"io"
	"log/slog"
	"net/http"
)

// Download downloads the track from the given URL and returns the path to the downloaded file
func (s *Soundcloud) Download(ctx context.Context, url string, info entity.TrackInfo) (string, error) {
	attr := []slog.Attr{
		slog.String("url", url),
		slog.String("method", "download"),
		slog.String("track_title", info.TrackTitle),
		slog.String("artist_name", info.ArtistName),
	}

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
	defer resp.Body.Close()

	// parse html
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("get artwork: %w", err)
	}

	streamURL, err := s.ConstructStreamURL(url, doc)
	if err != nil {
		return "", fmt.Errorf("construct stream url: %w", err)
	}

	var songName string
	if info.TrackTitle == "" {
		songName, err = s.GetTitle(doc)
		if err != nil {

			return "", fmt.Errorf("get title: %w", err)
		}
	} else {
		songName = info.TrackTitle
	}

	// Get artist
	var artistName string
	if info.ArtistName == "" {
		artistName, err = s.GetArtist(doc, songName)
		if err != nil {
			return "", fmt.Errorf("get artist: %w", err)
		}
	} else {
		artistName = info.ArtistName
	}

	// Get the artwork
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
	mp3Module, err := mp3.NewModule(s.proxyUrl, s.downloadPath, songName)
	if err != nil {
		return "", fmt.Errorf("mp3 new module: %w", err)
	}

	path, err := mp3Module.Merge(audioResp.URL)
	if err != nil {
		return "", fmt.Errorf("merge: %w", err)
	}

	var artworkResp *http.Response
	var image []byte

	artworkResp, err = http.Get(artwork)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "get artwork", logger.AppendErrorToLogs(attr, err)...)
	}

	if err == nil {
		defer artworkResp.Body.Close()
		image, err = io.ReadAll(artworkResp.Body)
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "read artwork body", logger.AppendErrorToLogs(attr, err)...)
		}
	}

	err = mp3Module.SetTitleArtistCoverImage(path, songName, artistName, image)
	if err != nil {
		s.logger.LogAttrs(ctx, slog.LevelError, "set cover image", logger.AppendErrorToLogs(attr, err)...)
	}

	return path, nil
}
