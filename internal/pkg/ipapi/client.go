package ipapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type Client struct {
	key    string
	client *http.Client
}

func NewClient(key string, client *http.Client) *Client {
	return &Client{key, client}
}

type LookupResult struct {
	IP          string `json:"ip"`
	Type        string `json:"type"`
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
}

func (c *Client) LookupIP(ip string) (*LookupResult, error) {
	queryURL := fmt.Sprintf("http://api.ipapi.com/%s?access_key=%s&format=1", ip, c.key)
	response, err := c.client.Get(queryURL)
	if err != nil {
		return nil, errors.Wrap(err, "error querying ipapi")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("ipapi bad status code %d", response.StatusCode)
	}

	var result LookupResult
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "error decoding ipapi response")
	}

	return &result, nil
}
