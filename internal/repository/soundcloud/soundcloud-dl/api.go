/*
This file contains all the soundcloud-dl api routes with params needed for this project.
*/
package soundcloud_dl

import (
	"net/url"
	"strings"
)

var (
	BaseApiUrl = "https://api-v2.soundcloud.com/"

	// get all info about the track through the url
	ResolveApiUrl = "https://api-widget.soundcloud.com/resolve?"

	// to search for tracks
	SearchTrackApiUrl = "https://api-v2.soundcloud.com/search/tracks?"

	TracksApiUrl = "https://api-v2.soundcloud.com/tracks?"
)

// GetTrackInfoAPIUrl resolve the given url: (return info about it).
func GetTrackInfoAPIUrl(urlx string, clientId string) string {
	v := url.Values{}

	// setting all the query params
	v.Set("url", urlx)
	v.Set("format", "json")
	v.Set("client_id", clientId)

	encodedUrl := ResolveApiUrl + v.Encode()

	return encodedUrl
}

func GetTracksByIdsApiUrl(ids []string, clientId string) string {

	v := url.Values{}

	// setting all the query params
	v.Set("client_id", clientId)
	v.Set("ids", strings.Join(ids, ","))

	encodedUrl := TracksApiUrl + v.Encode()

	return encodedUrl

}
