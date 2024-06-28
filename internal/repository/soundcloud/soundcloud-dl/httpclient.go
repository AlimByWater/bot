package soundcloud_dl

import (
	"io/ioutil"
	"net/http"
)

// make a GET request and return some info about the request
func Get(url1 string) (int, []byte, error) {
	resp, err := http.Get(url1)

	if err != nil {
		return -1, nil, err
	}

	// read the response body
	bodyBytes, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if err != nil {
		return resp.StatusCode, nil, err
	}

	return resp.StatusCode, bodyBytes, nil
}
