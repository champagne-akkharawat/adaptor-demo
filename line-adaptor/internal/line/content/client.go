package content

import "net/http"

// BaseURL is the LINE content API base. Overridable in tests.
var BaseURL = "https://api-data.line.me/v2/bot/message"

type Client struct {
	accessToken string
	http        *http.Client
}

func New(accessToken string) *Client {
	return &Client{accessToken: accessToken, http: &http.Client{}}
}
