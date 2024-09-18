package soundcloudV2

import (
	"bytes"
	"elysium/internal/entity"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/antchfx/htmlquery"
	"github.com/bogem/id3v2"
	"golang.org/x/net/html"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

func (m *Module) GetClientId(url string) (string, error) {

	if url == "" {
		url = "https://soundcloud.com/ahmed-yehia0"
	}

	statusCode, bodyData, err := m.Get(url)

	if err != nil {
		return "", fmt.Errorf("get %s: %w", url, err)
	}
	if statusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", statusCode)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyData))

	// find the last src under the body
	apiurl, exists := doc.Find("body > script").Last().Attr("src")
	if !exists {
		return "", fmt.Errorf("src doesn't exists")
	}

	// making a GET request to find the client_id
	resp, err := m.httpClient.Get(apiurl)
	if err != nil {
		return "", fmt.Errorf("something went wrong while requesting: %w", err)
	}

	// reading the body
	bodyData, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	defer resp.Body.Close()

	// search for the client_id
	pattern := ",client_id:\"([^\"]*?.[^\"]*?)\""
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(string(bodyData), 1)

	return matches[0][1], nil
}

func (m *Module) Get(u string) (int, []byte, error) {

	resp, err := m.httpClient.Get(u)

	if err != nil {
		return -1, nil, err
	}

	// read the response body
	bodyBytes, err := io.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, bodyBytes, nil
}

func (m *Module) GetTrackInfoAPIUrl(urlx string, clientId string) string {
	v := url.Values{}

	// setting all the query params
	v.Set("url", urlx)
	v.Set("format", "json")
	v.Set("client_id", clientId)

	encodedUrl := entity.ResolveApiUrl + v.Encode()

	return encodedUrl
}

func (m *Module) GetSoundMetaData(apiUrl string) (*entity.SoundData, error) {

	statusCode, body, err := m.Get(apiUrl)

	if err != nil || statusCode != http.StatusOK {
		return nil, fmt.Errorf("error while requesting %s , status: %d; error : %v", apiUrl, statusCode, err)
	}

	var sound entity.SoundData
	err = json.Unmarshal(body, &sound)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling the json: %w", err)
	}

	return &sound, nil
}

func (m *Module) GetFormattedDL(track *entity.SoundData, clientId string) []entity.DownloadTrack {

	ext := "mp3" // the default extension type
	tracks := make([]entity.DownloadTrack, 0)
	data := track.Transcodes.Transcodings
	var wg sync.WaitGroup

	for _, tcode := range data {
		wg.Add(1)
		go func(tcode entity.Transcode) {
			defer wg.Done()

			u := tcode.ApiUrl + "?client_id=" + clientId
			statusCode, body, err := m.Get(u)
			if err != nil && statusCode != http.StatusOK {
				return
			}
			q := mapQuality(tcode.ApiUrl, tcode.Format.MimeType)
			if q == "high" {
				ext = "ogg"
			}
			mediaUrl := entity.Media{}
			err = mediaUrl.UnmarshalJSON(body)
			//dec := json.NewDecoder(bytes.NewReader(body))
			if err != nil {
				m.logger.Error("Error unmarshal media json", slog.String("err", err.Error()), slog.String("u", u))
				return
			}
			tmpTrack := entity.DownloadTrack{
				Url:       mediaUrl.Url,
				Quality:   q,
				SoundData: track,
				Ext:       ext,
			}
			tracks = append(tracks, tmpTrack)

		}(tcode)
	}
	wg.Wait()
	return tracks
}

// check if the trackUrl is mp3:progressive or ogg:hls
func mapQuality(url string, format string) string {
	tmp := strings.Split(url, "/")
	if tmp[len(tmp)-1] == "hls" && strings.HasPrefix(format, "audio/ogg") {
		return "high"
	} else if tmp[len(tmp)-1] == "hls" && strings.HasPrefix(format, "audio/mpeg") {
		return "medium"
	}
	return "low"
}

func getTrack(downloadTracks []entity.DownloadTrack) entity.DownloadTrack {
	var defaultQuality = "medium"
	// show available download options
	qualities := getQualities(downloadTracks)
	defaultQuality = getHighestQuality(qualities)

	return chooseTrackDownload(downloadTracks, defaultQuality)

}

func chooseTrackDownload(tracks []entity.DownloadTrack, target string) entity.DownloadTrack {
	for _, track := range tracks {
		if track.Quality == target {
			return track
		}
	}
	return tracks[0]
}

// get all the available qualities inside the track
// used to choose a track to download based on the target quality
func getQualities(tracks []entity.DownloadTrack) []string {
	qualities := make([]string, 0)
	for _, track := range tracks {
		// check the default one
		qualities = append(qualities, track.Quality)
	}
	return qualities
}

func getHighestQuality(qualities []string) string {
	allQualities := []string{"high", "medium", "low"}
	var in = func(a string, list []string) bool {
		for _, b := range list {
			if b == a {
				return true
			}
		}
		return false
	}

	for _, q := range allQualities {
		if in(q, qualities) {
			return q
		}
	}
	return ""
}

func SetTitleArtistCoverImage(filepath, title, artist string, image []byte) error {
	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		return fmt.Errorf("error while opening mp3 file: %w", err)
	}

	if len(image) != 0 {
		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingUTF8,
			MimeType:    "image/jpeg",
			PictureType: id3v2.PTOther,
			Description: "Album cover",
			Picture:     image,
		}

		tag.AddAttachedPicture(pic)
	}

	textFrame := id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF16,
		Text:     title,
	}

	tag.AddFrame(tag.CommonID("Title"), textFrame)

	textFrame = id3v2.TextFrame{
		Encoding: id3v2.EncodingUTF16,
		Text:     artist,
	}

	tag.AddFrame(tag.CommonID("Artist"), textFrame)

	return tag.Save()

}

// getTitle will return the title of the song
func getTitle(doc *html.Node) (string, error) {
	// XPath query
	titlePath := "//meta[@property='og:title']/@content"

	// Query the document for the title node
	nodes, err := htmlquery.QueryAll(doc, titlePath)
	if err != nil {
		return "", fmt.Errorf("error executing XPath query: %w", err)
	}

	// Check if any nodes were found
	if len(nodes) > 0 {
		// Extract the content from the first node
		title := htmlquery.InnerText(nodes[0])
		return title, nil
	}

	return "", fmt.Errorf("no title found")
}

// getArtist will return the artist of the song
func getArtist(doc *html.Node, songTitle string) (string, error) {
	// XPath query to find the title
	titlePath := "//title"

	// Query the document for the title node
	node := htmlquery.FindOne(doc, titlePath)
	if node != nil {
		// Extract the content from the node
		title := htmlquery.InnerText(node)

		t := strings.SplitAfter(title, songTitle+" by ")
		if len(t) > 1 {
			return t[1], nil
		}

	}

	return "", fmt.Errorf("no artist found")
}
