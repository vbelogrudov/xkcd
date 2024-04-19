package xkcd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const requestTimeout = 10 * time.Second
const formatURL = "%s/%d/info.0.json"

var ErrNotFound = fmt.Errorf("not found")

type Comics struct {
	ID         int    `json:"num"`
	Title      string `json:"title"`
	Image      string `json:"img"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
}

type XKCD struct {
	URL string
	http.Client
}

func New(URL string) *XKCD {
	return &XKCD{
		URL:    URL,
		Client: http.Client{Timeout: requestTimeout},
	}
}

func (x XKCD) Get(ID int) (Comics, error) {
	URL := fmt.Sprintf(formatURL, x.URL, ID)
	response, err := x.Client.Get(URL)
	if err != nil {
		return Comics{}, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusNotFound:
		return Comics{}, ErrNotFound
	case http.StatusOK: // ok!
	default:
		return Comics{}, fmt.Errorf("xkcd returned %s", response.Status)
	}

	var comics Comics
	err = json.NewDecoder(response.Body).Decode(&comics)
	return comics, err
}
