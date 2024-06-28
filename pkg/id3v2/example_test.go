package id3v2_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"sync"
)

func Example() {
	// Open file and parse tag in it.
	tag, err := Open("file.mp3", Options{Parse: true})
	if err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}
	defer tag.Close()

	// Read frames.
	fmt.Println(tag.Artist())
	fmt.Println(tag.Title())

	// Set simple text frames.
	tag.SetArtist("Artist")
	tag.SetTitle("Title")

	// Set comment frame.
	comment := CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    "eng",
		Description: "My opinion",
		Text:        "Very good song",
	}
	tag.AddCommentFrame(comment)

	// Write tag to file.
	if err = tag.Save(); err != nil {
		log.Fatal("Error while saving a tag: ", err)
	}
}

func Example_concurrent() {
	tagPool := sync.Pool{New: func() interface{} { return NewEmptyTag() }}

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			tag := tagPool.Get().(*Tag)
			defer tagPool.Put(tag)

			file, err := os.Open("file.mp3")
			if err != nil {
				log.Fatal("Error while opening file:", err)
			}
			defer file.Close()

			if err := tag.Reset(file, Options{Parse: true}); err != nil {
				log.Fatal("Error while reseting tag to file:", err)
			}

			fmt.Println(tag.Artist() + " - " + tag.Title())
		}()
	}
	wg.Wait()
}

func ExampleTag_AddAttachedPicture() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	artwork, err := ioutil.ReadFile("artwork.jpg")
	if err != nil {
		log.Fatal("Error while reading artwork file", err)
	}

	pic := PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Description: "Front cover",
		Picture:     artwork,
	}
	tag.AddAttachedPicture(pic)
}

func ExampleTag_AddCommentFrame() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	comment := CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    "eng",
		Description: "My opinion",
		Text:        "Very good song",
	}
	tag.AddCommentFrame(comment)
}

func ExampleTag_AddUnsynchronisedLyricsFrame() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	uslt := UnsynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          "ger",
		ContentDescriptor: "Deutsche Nationalhymne",
		Lyrics:            "Einigkeit und Recht und Freiheit...",
	}
	tag.AddUnsynchronisedLyricsFrame(uslt)
}

func ExampleTag_GetFrames() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	for _, f := range pictures {
		pic, ok := f.(PictureFrame)
		if !ok {
			log.Fatal("Couldn't assert picture frame")
		}
		// Do something with picture frame.
		// For example, print description of picture frame:
		fmt.Println(pic.Description)
	}
}

func ExampleTag_GetLastFrame() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	bpmFramer := tag.GetLastFrame(tag.CommonID("BPM"))
	if bpmFramer != nil {
		bpm, ok := bpmFramer.(TextFrame)
		if !ok {
			log.Fatal("Couldn't assert bpm frame")
		}
		fmt.Println(bpm.Text)
	}
}

func ExampleCommentFrame_get() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	comments := tag.GetFrames(tag.CommonID("Comments"))
	for _, f := range comments {
		comment, ok := f.(CommentFrame)
		if !ok {
			log.Fatal("Couldn't assert comment frame")
		}

		// Do something with comment frame.
		// For example, print the text:
		fmt.Println(comment.Text)
	}
}

func ExampleCommentFrame_add() {
	tag := NewEmptyTag()
	comment := CommentFrame{
		Encoding:    EncodingUTF8,
		Language:    "eng",
		Description: "My opinion",
		Text:        "Very good song",
	}
	tag.AddCommentFrame(comment)
}

func ExamplePictureFrame_get() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	pictures := tag.GetFrames(tag.CommonID("Attached picture"))
	for _, f := range pictures {
		pic, ok := f.(PictureFrame)
		if !ok {
			log.Fatal("Couldn't assert picture frame")
		}

		// Do something with picture frame.
		// For example, print the description:
		fmt.Println(pic.Description)
	}
}

func ExamplePictureFrame_add() {
	tag := NewEmptyTag()
	artwork, err := ioutil.ReadFile("artwork.jpg")
	if err != nil {
		log.Fatal("Error while reading artwork file", err)
	}

	pic := PictureFrame{
		Encoding:    EncodingUTF8,
		MimeType:    "image/jpeg",
		PictureType: PTFrontCover,
		Description: "Front cover",
		Picture:     artwork,
	}
	tag.AddAttachedPicture(pic)
}

func ExampleTextFrame_get() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	tf := tag.GetTextFrame(tag.CommonID("Mood"))
	fmt.Println(tf.Text)
}

func ExampleTextFrame_add() {
	tag := NewEmptyTag()
	textFrame := TextFrame{
		Encoding: EncodingUTF8,
		Text:     "Happy",
	}
	tag.AddFrame(tag.CommonID("Mood"), textFrame)
}

func ExampleUnsynchronisedLyricsFrame_get() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	uslfs := tag.GetFrames(tag.CommonID("Unsynchronised lyrics/text transcription"))
	for _, f := range uslfs {
		uslf, ok := f.(UnsynchronisedLyricsFrame)
		if !ok {
			log.Fatal("Couldn't assert USLT frame")
		}

		// Do something with USLT frame.
		// For example, print the lyrics:
		fmt.Println(uslf.Lyrics)
	}
}

func ExampleUnsynchronisedLyricsFrame_add() {
	tag := NewEmptyTag()
	uslt := UnsynchronisedLyricsFrame{
		Encoding:          EncodingUTF8,
		Language:          "ger",
		ContentDescriptor: "Deutsche Nationalhymne",
		Lyrics:            "Einigkeit und Recht und Freiheit...",
	}
	tag.AddUnsynchronisedLyricsFrame(uslt)
}

func ExamplePopularimeterFrame_add() {
	tag := NewEmptyTag()

	popmFrame := PopularimeterFrame{
		Email:   "foo@bar.com",
		Rating:  128,
		Counter: big.NewInt(10000000000000000),
	}
	tag.AddFrame(tag.CommonID("Popularimeter"), popmFrame)
}

func ExamplePopularimeterFrame_get() {
	tag, err := Open("file.mp3", Options{Parse: true})
	if tag == nil || err != nil {
		log.Fatal("Error while opening mp3 file: ", err)
	}

	f := tag.GetLastFrame(tag.CommonID("Popularimeter"))
	popm, ok := f.(PopularimeterFrame)
	if !ok {
		log.Fatal("Couldn't assert POPM frame")
	}

	// do something with POPM Frame
	fmt.Printf("Email: %s, Rating: %d, Counter: %d", popm.Email, popm.Rating, popm.Counter)
}
