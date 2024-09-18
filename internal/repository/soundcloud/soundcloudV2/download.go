package soundcloudV2

import (
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/grafov/m3u8"
	"golang.org/x/net/html"
	"io"
	"log/slog"
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

type Module struct {
	httpClient *http.Client
	logger     *slog.Logger
}

func NewModule(httpClient *http.Client, log *slog.Logger) *Module {
	return &Module{
		logger:     log.With(slog.String("module", "☁️ soundcloud-dl2 ")),
		httpClient: httpClient,
	}
}

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
func (m *Module) downloadSeg(wg *sync.WaitGroup, segmentURI string, file *os.File) error {
	defer wg.Done()
	resp, err := m.httpClient.Get(segmentURI)

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

func (m *Module) getSegments(body io.Reader) []string {
	segments := make([]string, 0)
	pl, listType, err := m3u8.DecodeFrom(body, true)

	if err != nil {
		return nil
	}

	switch listType {
	case m3u8.MEDIA:
		mediapl := pl.(*m3u8.MediaPlaylist)
		for _, segment := range mediapl.Segments {
			if segment == nil {
				continue
			}
			segments = append(segments, segment.URI)
		}
	}
	return segments
}

func (m *Module) DownloadM3u8(filepath string, segments []string) error {

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}

	// the go routine now
	var wg sync.WaitGroup

	for _, segment := range segments {
		wg.Add(1)
		err = m.downloadSeg(&wg, segment, file)
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
func (m *Module) Download(track entity.DownloadTrack, dlpath string) (string, error) {
	// TODO: Prompt Y/N if the file exists and rename by adding _<random/date>.<ext>
	p := path.Join(dlpath, strconv.FormatInt(time.Now().Unix(), 10)+"."+track.Ext)
	//p, err := validateDownload(dlpath, trackName)
	//if err != nil {
	//	return "", fmt.Errorf("validate download: %w", err)
	//}

	// check if the track is hls
	if track.Quality != "low" {

		resp, err := m.httpClient.Get(track.Url)
		if err != nil {
			return "", fmt.Errorf("get track: %w", err)
		}
		defer resp.Body.Close()

		segments := m.getSegments(resp.Body)
		err = m.DownloadM3u8(p, segments)
		if err != nil {
			return "", fmt.Errorf("download m3u8: %w", err)
		}

	} else {
		resp, err := m.httpClient.Get(track.Url)

		if err != nil {
			return "", fmt.Errorf("get track: %w", err)
		}
		defer resp.Body.Close()

		// check if the file exists
		f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
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
		newPath := strings.Replace(p, ".ogg", ".mp3", 1)
		cmd := exec.Command("ffmpeg", "-i", p, newPath)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("ffmpeg: %w", err)
		}

		os.Remove(p)
		p = newPath
	}

	return p, nil
}

func (m *Module) GetLongUrl(url string) (string, error) {
	resp, err := m.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("get long url: %w", err)
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("new document from reader: %w", err)
	}

	longUrl := doc.Find("meta[property='og:url']").First().AttrOr("content", "")
	if longUrl == "" {
		return "", fmt.Errorf("long url is empty; status code: %d", resp.StatusCode)
	}

	return longUrl, nil
}

func (m *Module) DownloadByUrl(url string, dlpath string, info entity.TrackInfo) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is empty")
	}
	var err error

	if strings.Contains(url, "on.soundcloud.com") {
		url, err = m.GetLongUrl(url)
		if err != nil {
			return "", fmt.Errorf("get long url: %w", err)
		}
	}

	clientID, err := m.GetClientId(url)
	if err != nil {
		return "", fmt.Errorf("get client id: %w", err)
	}

	if clientID == "" {
		return "", fmt.Errorf("client id is empty")
	}

	apiUrl := m.GetTrackInfoAPIUrl(url, clientID)
	soundData, err := m.GetSoundMetaData(apiUrl)
	if err != nil {
		return "", fmt.Errorf("get sound data: %w", err)
	}

	if soundData == nil {
		return "", fmt.Errorf("sound data is empty")
	}

	downloadTracks := m.GetFormattedDL(soundData, clientID)
	if len(downloadTracks) == 0 {
		return "", fmt.Errorf("download tracks is empty")
	}

	track := getTrack(downloadTracks)
	filePath, err := m.Download(track, dlpath)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}

	artWorkurl := strings.Replace(track.SoundData.ArtworkUrl, "large", "t500x500", 1)

	// fetching the data
	statusCode, imgdata, artErr := m.Get(artWorkurl)
	if artErr != nil || statusCode != http.StatusOK {
		imgdata = nil
	}

	if info.TrackTitle == "" || info.ArtistName == "" {
		resp, err := m.httpClient.Get(url)
		if err != nil {
			return "", fmt.Errorf("get so: %w", err)
		}
		defer resp.Body.Close()

		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			return "", fmt.Errorf("get artwork: %w", err)
		}

		info.TrackTitle, err = getTitle(doc)
		if err != nil {

			return "", fmt.Errorf("get title: %w", err)
		}

		// Get artist
		info.ArtistName, err = getArtist(doc, info.TrackTitle)
		if err != nil {
			return "", fmt.Errorf("get artist: %w", err)
		}

		info.ArtistName = strings.Replace(info.ArtistName, " | Listen online for free on SoundCloud", "", 1)

		if artErr != nil || imgdata == nil {
			artworkUrl, err := GetArtwork(doc)
			if err != nil {
				return "", fmt.Errorf("get artwork: %w", err)
			}

			artworkResp, err := m.httpClient.Get(artworkUrl)
			if err != nil {
				return "", fmt.Errorf("get artwork: %w", err)
			}
			defer artworkResp.Body.Close()
			imgdata, err = io.ReadAll(artworkResp.Body)
			if err != nil {
				return "", fmt.Errorf("read artwork body: %w", err)
			}
		}
	}
	err = SetTitleArtistCoverImage(filePath, info.TrackTitle, info.ArtistName, imgdata)
	if err != nil {
		return "", fmt.Errorf("set title artist cover image: %w", err)
	}

	return filePath, nil
}

func GetArtwork(doc *html.Node) (string, error) {
	// XPath query
	artworkPath := "//meta[@property='og:image']/@content"

	// Query the document for the artwork node
	nodes, err := htmlquery.QueryAll(doc, artworkPath)
	if err != nil {
		return "", fmt.Errorf("error executing XPath query: %w", err)
	}

	// Check if any nodes were found
	if len(nodes) > 0 {
		// Extract the content from the first node
		artwork := htmlquery.InnerText(nodes[0])

		return artwork, nil
	}

	return "", nil
}
