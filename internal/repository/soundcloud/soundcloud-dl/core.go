package soundcloud_dl

import (
	"elysium/internal/repository/soundcloud/soundcloud-dl/theme"
	"fmt"
	"strings"
)

var (
	defaultQuality = "medium"
	soundData      = &SoundData{}
)

func Sc(args []string, downloadPath string) (err error) {

	url := ""
	if len(args) > 0 {
		url = args[0]
	}

	if url != "" && !initValidations(url) {
		return nil
	}

	clientId, err := GetClientId(url)

	if clientId == "" {
		fmt.Println("Something went wrong while getting the Client Id!")
		//return
	}

	apiUrl := GetTrackInfoAPIUrl(url, clientId)
	soundData = GetSoundMetaData(apiUrl, url, clientId)
	if soundData == nil {
		fmt.Printf("%s URL : %s doesn't return a valid track. Track is publicly accessed ?", theme.Red("[+]"), theme.Magenta(url))
		return
	}

	fmt.Printf("%s %s found. Title : %s - Duration : %s\n", theme.Green("[+]"), strings.Title(soundData.Kind), theme.Magenta(soundData.Title), theme.Magenta(theme.FormatTime(soundData.Duration)))

	// check if the url is a playlist
	//if soundData.Kind == "playlist" {
	//	var wg sync.WaitGroup
	//	plDownloadTracks := getPlaylistDownloadTracks(soundData, clientId)
	//
	//	for _, dlT := range plDownloadTracks {
	//
	//		wg.Add(1)
	//
	//		go func(dlT []DownloadTrack) {
	//			defer wg.Done()
	//			// bestQuality is true to avoid prompting the user for quality choosing each time and speed up
	//			// TODO: get a single progress bar, this will require the use of "https://github.com/cheggaaa/pb" since the current pb doesn't support download pool (I think)
	//			t := getTrack(dlT)
	//			fp, err := Download(t, downloadPath)
	//
	//			// silent indication of already existing files
	//			if fp == "" {
	//				return
	//			}
	//			AddMetadata(t, fp)
	//
	//		}(dlT)
	//	}
	//	wg.Wait()
	//
	//	fmt.Printf("\n%s Playlist saved to : %s\n", theme.Green("[-]"), theme.Magenta(downloadPath))
	//	return
	//}

	downloadTracks := GetFormattedDL(soundData, clientId)

	track := getTrack(downloadTracks)
	filePath, err := Download(track, downloadPath)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	// add tags
	if filePath == "" {
		fmt.Printf("\n%s Track was already saved to : %s\n", theme.Green("[-]"), theme.Magenta(downloadPath))
		return
	}
	err = AddMetadata(track, filePath)
	if err != nil {
		return fmt.Errorf("error happend while adding tags to the track : %w", err)
	}
	fmt.Printf("\n%s Track saved to : %s\n", theme.Green("[-]"), theme.Magenta(filePath))
	return nil
}
