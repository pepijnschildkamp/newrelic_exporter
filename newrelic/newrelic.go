package newrelic

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/prometheus/common/log"
)

type transport struct {
	transport http.RoundTripper
	apiKey    string
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("X-Api-Key", t.apiKey)
	r.Header.Add("User-Agent", "newrelic_exporter;go")
	return t.transport.RoundTrip(r)
}

// A Client manages communication with the NewRelic API
type Client struct {
	baseURL *url.URL
	client  *http.Client
}

// NewClient returns an initialized NewRelic API client
func NewClient(apiURL, apiKey string) *Client {
	baseURL, err := url.Parse(apiURL)
	if err != nil {
		log.Fatalf("Failed to parse New Relic API URL: %v", err)
	}

	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: time.Second * 10,
			Transport: &transport{
				transport: http.DefaultTransport,
				apiKey:    apiKey,
			},
		},
	}
}

func (c *Client) newRequest(method, path string) (*http.Request, error) {
	url, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(method, url.String(), nil)
}

func (c *Client) do(req *http.Request, result interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.Warn("An error occured closing response body: ", err)
		}
	}()

	return resp, json.NewDecoder(resp.Body).Decode(result)
}
