package soundcloud

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"strings"
)

// GetTitle will return the title of the song
func (s *Soundcloud) GetTitle(doc *html.Node) (string, error) {
	// XPath query
	titlePath := "//meta[@property='og:title']/@content"

	// Query the document for the title node
	nodes, err := htmlquery.QueryAll(doc, titlePath)
	if err != nil {
		fmt.Println("Error executing XPath query:", err)
		return "", err
	}

	// Check if any nodes were found
	if len(nodes) > 0 {
		// Extract the content from the first node
		title := htmlquery.InnerText(nodes[0])
		return title, nil
	}

	return "", fmt.Errorf("no title found")
}

// GetArtist will return the artist of the song
func (s *Soundcloud) GetArtist(doc *html.Node, songTitle string) (string, error) {
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
		fmt.Println(t)
	}

	return "", fmt.Errorf("no artist found")
}
