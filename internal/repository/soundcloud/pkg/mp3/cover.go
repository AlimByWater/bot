package mp3

import (
	"fmt"

	"github.com/bogem/id3v2"
	"github.com/fatih/color"
)

// SetTitleArtistCoverImage sets the id3v2 cover image meta tag of a given .mp3 file
func SetTitleArtistCoverImage(filepath string, title, artist string, image []byte) {
	// replace empty cover image with SoundCloud artwork
	tag, err := id3v2.Open(filepath, id3v2.Options{Parse: true})
	if tag == nil || err != nil {
		red := color.New(color.FgRed).SprintFunc()
		fmt.Printf("%s Error while opening mp3 file: %s %s\n", red("[-]"), err, filepath)

	}

	tag.SetArtist(artist)
	tag.SetTitle(title)
	tag.SetAlbum(title)

	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTOther,
		Description: "Album cover",
		Picture:     image,
	}

	tag.AddAttachedPicture(pic)

	tag.Save()
}
