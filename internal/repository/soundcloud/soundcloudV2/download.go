package soundcloudV2

import (
	"arimadj-helper/internal/entity"
	"errors"
	"fmt"
	m "github.com/grafov/m3u8"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

// validate the given path
// and check if the file already exists or not
func fileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

// extract the urls of the individual segment and then steam/download.
func downloadSeg(wg *sync.WaitGroup, segmentURI string, file *os.File) error {
	defer wg.Done()
	resp, err := http.Get(segmentURI)

	if err != nil {
		return fmt.Errorf("get segment: %w", err)
	}

	defer resp.Body.Close()

	_, err = io.Copy(io.MultiWriter(file), resp.Body)

	if err != nil {
		return fmt.Errorf("copy segment: %w", err)
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

func DownloadM3u8(filepath string, segments []string) error {

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	// the go routine now
	var wg sync.WaitGroup

	for _, segment := range segments {
		wg.Add(1)
		err = downloadSeg(&wg, segment, file)
		if err != nil {
			return fmt.Errorf("download seg: %w", err)
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
		return "", fmt.Errorf("file dont exists: %w", err)
	}
	return path, nil
}

// download the track
func Download(track DownloadTrack, dlpath string) (string, error) {
	// TODO: Prompt Y/N if the file exists and rename by adding _<random/date>.<ext>
	path := path.Join(dlpath, strconv.FormatInt(time.Now().Unix(), 10)+"."+track.Ext)
	//path, err := validateDownload(dlpath, trackName)
	//if err != nil {
	//	return "", fmt.Errorf("validate download: %w", err)
	//}

	// check if the track is hls
	if track.Quality != "low" {

		resp, err := http.Get(track.Url)
		if err != nil {
			return "", fmt.Errorf("get track: %w", err)
		}
		defer resp.Body.Close()

		segments := getSegments(resp.Body)
		err = DownloadM3u8(path, segments)
		if err != nil {
			return "", fmt.Errorf("download m3u8: %w", err)
		}

	} else {
		resp, err := http.Get(track.Url)

		if err != nil {
			return "", fmt.Errorf("get track: %w", err)
		}
		defer resp.Body.Close()

		// check if the file exists
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return "", fmt.Errorf("open file: %w", err)
		}
		defer f.Close()

		_, err = io.Copy(io.MultiWriter(f), resp.Body)
		if err != nil {
			return "", fmt.Errorf("copy: %w", err)
		}
	}

	if track.Ext == "ogg" {
		newPath := strings.Replace(path, ".ogg", ".mp3", 1)
		cmd := exec.Command("ffmpeg", "-i", path, newPath)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("ffmpeg: %w", err)
		}

		os.Remove(path)
		path = newPath
	}

	return path, nil
}

func DownloadByUrl(url string, dlpath string, info entity.TrackInfo) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is empty")
	}

	clientID, err := GetClientId(url)
	if err != nil {
		return "", fmt.Errorf("get client id: %w", err)
	}

	if clientID == "" {
		return "", fmt.Errorf("client id is empty")
	}

	apiUrl := GetTrackInfoAPIUrl(url, clientID)
	soundData, err := GetSoundMetaData(apiUrl)
	if err != nil {
		return "", fmt.Errorf("get sound data: %w", err)
	}

	if soundData == nil {
		return "", fmt.Errorf("sound data is empty")
	}

	downloadTracks := GetFormattedDL(soundData, clientID)
	if len(downloadTracks) == 0 {
		return "", fmt.Errorf("download tracks is empty")
	}

	track := getTrack(downloadTracks)
	filePath, err := Download(track, dlpath)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	url = strings.Replace(track.SoundData.ArtworkUrl, "large", "t500x500", 1)

	// fetching the data
	statusCode, imgdata, err := Get(url)
	if err != nil || statusCode != http.StatusOK {
		return "", fmt.Errorf("get artwork: %w", err)
	}

	err = SetTitleArtistCoverImage(filePath, info.TrackTitle, info.ArtistName, imgdata)
	if err != nil {
		return "", fmt.Errorf("set title artist cover image: %w", err)
	}

	return filePath, nil
}
