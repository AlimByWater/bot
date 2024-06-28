package soundcloud_dl

import (
	"arimadj-helper/internal/repository/soundcloud/soundcloud-dl/theme"
	"fmt"
	"sync"
)

func initValidations(url string) bool {

	fmt.Printf("%s Validating the URL : %s\n", theme.Yellow("[+]"), theme.Magenta(url))

	// check if the url is valid
	if !IsValidUrl(url) {
		fmt.Printf("%s The Url : %s isn't a valid soundcloud URL\n", theme.Red("[+]"), theme.Magenta(url))
		return false
	}

	fmt.Printf("%s URL is Valid! \n", theme.Green("[+]"))
	fmt.Printf("%s Checking The Track on Soundcloud. \n", theme.Red("[+]"))

	return true

}

// TEMP: Just for now, return the quality
// the default quality is just mp3, highest is ogg
// if the quality doesn't exist return the first one
func chooseTrackDownload(tracks []DownloadTrack, target string) DownloadTrack {
	for _, track := range tracks {
		if track.Quality == target {
			return track
		}
	}
	return tracks[0]
}

// get all the available qualities inside the track
// used to choose a track to download based on the target quality
func getQualities(tracks []DownloadTrack) []string {
	qualities := make([]string, 0)
	for _, track := range tracks {
		// check the default one
		qualities = append(qualities, track.Quality)
	}
	return qualities
}

func getHighestQuality(qualities []string) string {
	allQualities := []string{"medium", "low"}
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

func getPlaylistDownloadTracks(soundData *SoundData, clientId string) [][]DownloadTrack {

	var wg sync.WaitGroup
	listDownloadTracks := make([][]DownloadTrack, 0)

	playlistTracks := GetPlaylistTracks(soundData, clientId)

	for i, t := range playlistTracks {
		wg.Add(1)

		go func(t SoundData) {
			defer wg.Done()
			dlTrack := GetFormattedDL(&t, clientId)
			listDownloadTracks = append(listDownloadTracks, dlTrack)
		}(t)
		fmt.Printf("%s  %v - %s \n", theme.Green("[+]"), theme.Red(i+1), theme.Magenta(t.Title))
	}
	wg.Wait()
	return listDownloadTracks
}

// get a final track to be downloaded
// if bestQuality is false it will prompt the user to choose a quality
func getTrack(downloadTracks []DownloadTrack) DownloadTrack {

	// show available download options
	qualities := getQualities(downloadTracks)
	defaultQuality = getHighestQuality(qualities)

	return chooseTrackDownload(downloadTracks, defaultQuality)

}
