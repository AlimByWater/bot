package soundcloud

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
)

// Unfortunately, SoundCloud does not inject the client ID into the page source unless the request is made from a browser. ( Javascript is enabled )
// This is a workaround to get the client ID from the JS file that is injected into the page source.

// GetClientID returns a new generated client_id when a request is made to SoundCloud's API
func (s *Soundcloud) GetClientID2() (string, error) {
	var clientID string

	// this is the JS file that is injected into the page source
	// this can always change at some point, so we have to keep an eye on it
	resp, err := s.Client.Get("https://a-v2.sndcdn.com/assets/2-fbfac8ab.js")
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("failed to read response body: %v", err)
	}

	// regex to find the client_id
	re := regexp.MustCompile(`client_id\s*:\s*['"]([^'"]+)['"]`)
	matches := re.FindSubmatch(body)

	if len(matches) > 1 {
		// Found a client_id
		clientID = string(matches[1])
	} else {
		log.Println("client_id not found")
		return "", fmt.Errorf("client_id not found")
	}

	return clientID, nil
}

func (s *Soundcloud) GetClientID(trackURl string) (string, error) {
	resp, err := s.Client.Get(trackURl)
	if err != nil {
		return "", fmt.Errorf("client get: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read all body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("new doc from reader: %w", err)
	}

	// find the last src under the body
	apiurl, exists := doc.Find("body > script").Last().Attr("src")
	if !exists {
		return "", fmt.Errorf("src don't exist")
	}

	// making a GET request to find the client_id
	resp, err = http.Get(apiurl)
	if err != nil {
		return "", fmt.Errorf("http get %s: %w", apiurl, err)
	}

	// reading the body
	bodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read apiurl body: %w", err)
	}

	defer resp.Body.Close()

	// search for the client_id
	pattern := ",client_id:\"([^\"]*?.[^\"]*?)\""
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(string(bodyData), 1)

	return matches[0][1], nil
}
