// adding tags to the track after downloading it.
package soundcloud_dl

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bogem/id3v2"
)

func AddMetadata(track DownloadTrack, filePath string) error {
	t500 := "t500x500" // for getting a higher res img
	imgBytes := make([]byte, 0)

	// check for artist thing
	if track.SoundData.ArtworkUrl != "" {
		url := strings.Replace(track.SoundData.ArtworkUrl, "large", t500, 1)

		// fetching the data
		resp, err := http.Get(url)

		if err != nil || resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http get: %w", err)
		}

		// read the response body
		bodyBytes, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return fmt.Errorf("read resp body: %w", err)
		}

		imgBytes = bodyBytes
	}

	tag, err := id3v2.Open(filePath, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	// setting metadata
	tag.SetTitle(track.SoundData.Title)
	tag.SetGenre(track.SoundData.Genre)
	tag.SetYear(track.SoundData.CreatedAt)

	// extracting the usr
	artistName := strings.Split(track.SoundData.PermalinkUrl, "/")
	tag.SetArtist(artistName[3])

	if imgBytes != nil {
		tag.AddAttachedPicture(
			id3v2.PictureFrame{
				Encoding:    id3v2.EncodingUTF8,
				MimeType:    "image/jpeg",
				Picture:     imgBytes,
				Description: track.SoundData.Description, // well, coz why not :D
			},
		)
	}
	if err = tag.Save(); err != nil {
		return err
	}
	return nil

}
