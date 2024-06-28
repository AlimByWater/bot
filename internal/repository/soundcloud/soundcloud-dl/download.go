// download the soundcloud-dl tracks
package soundcloud_dl

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sync"

	m "github.com/grafov/m3u8"
	bar "github.com/schollz/progressbar/v3"
)

// +
// expand the given path ~/Desktop to the current logged in user /home/<username>/Desktop
func expandPath(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil

}

// +
// validate the given path
// and check if the file already exists or not
func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// extract the urls of the individual segment and then steam/download.
func downloadSeg(wg *sync.WaitGroup, segmentURI string, file *os.File, dlbar *bar.ProgressBar) error {
	defer wg.Done()
	resp, err := http.Get(segmentURI)

	if err != nil {
		return fmt.Errorf("http get: %w", err)
	}

	defer resp.Body.Close()

	// append to the file
	if dlbar == nil {
		_, err = io.Copy(io.MultiWriter(file), resp.Body)
	} else {
		_, err = io.Copy(io.MultiWriter(file, dlbar), resp.Body)
	}

	if err != nil {
		return fmt.Errorf("io copy: %w", err)
	}

	return nil
}

func getSegments(body io.Reader) []string {
	segments := make([]string, 0)
	pl, listType, err := m.DecodeFrom(body, true)

	if err != nil {
		return nil
	}

	switch listType {
	case m.MEDIA:
		mediapl := pl.(*m.MediaPlaylist)
		for _, segment := range mediapl.Segments {
			if segment == nil {
				continue
			}
			segments = append(segments, segment.URI)
		}
	}
	return segments
}

// DownloadM3u8 using the goroutine to download each segment concurrently and wait till all finished
func DownloadM3u8(filepath string, dlbar *bar.ProgressBar, segments []string) error {

	file, _ := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)

	// the go routine now
	var wg sync.WaitGroup

	for _, segment := range segments {
		wg.Add(1)
		err := downloadSeg(&wg, segment, file, dlbar)
		if err != nil {
			return fmt.Errorf("downloadSeg: %w", err)
		}
	}

	return nil
}

// before download validation
// return the path if everything is alright.
func validateDownload(dlpath string, trackName string) (string, error) {

	testPath := path.Join(dlpath, trackName)
	path, err := expandPath(testPath)

	// TODO: handle all different kind of errors
	if fileExists(path) || err != nil {
		return "", fmt.Errorf("expand path: %w", err)
	}
	return path, nil
}

// download the track
func Download(track DownloadTrack, dlpath string) (string, error) {
	// TODO: Prompt Y/N if the file exists and rename by adding _<random/date>.<ext>
	trackName := track.SoundData.Title + "[" + track.Quality + "]." + track.Ext
	finalPath, err := validateDownload(dlpath, trackName)
	if err != nil {
		return "", fmt.Errorf("validate download: %w", err)
	}

	resp, err := http.Get(track.Url)
	if err != nil {
		return "", fmt.Errorf("http get track url: %w", err)
	}
	defer resp.Body.Close()

	// check if the track is hls
	if track.Quality != "low" {

		dlbar := bar.DefaultBytes(
			resp.ContentLength,
			"Downloading",
		)
		segments := getSegments(resp.Body)
		err := DownloadM3u8(finalPath, dlbar, segments)
		if err != nil {
			return "", fmt.Errorf("downloadm3u8: %w", err)
		}

		return finalPath, nil
	}

	// check if the file exists
	f, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return "", fmt.Errorf("os open file: %w", err)
	}

	bar := bar.DefaultBytes(
		resp.ContentLength,
		"Downloading",
	)

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return "", fmt.Errorf("io copy: %w", err)
	}

	return finalPath, nil
}
