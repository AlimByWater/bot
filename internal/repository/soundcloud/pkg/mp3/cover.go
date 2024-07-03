package mp3

import (
	"fmt"
	"github.com/bogem/id3v2"
)

// SetTitleArtistCoverImage sets the title, artist and cover image of the mp3 file
func (m *Module) SetTitleArtistCoverImage(filepath, title, artist string, image []byte) error {
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
