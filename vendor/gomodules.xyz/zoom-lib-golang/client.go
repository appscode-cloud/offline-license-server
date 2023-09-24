package zoom

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/go-querystring/query"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	apiURI     = "api.zoom.us"
	apiVersion = "/v2"
)

var (
	// Debug causes debugging message to be printed, using the log package,
	// when set to true
	Debug = false

	// APIKey is a package-wide API key, used when no client is instantiated
	APIKey string

	// APISecret is a package-wide API secret, used when no client is instantiated
	APISecret string

	defaultClient *Client
)

// Client is responsible for making API requests
type Client struct {
	cfg      clientcredentials.Config
	Timeout  time.Duration // set to value > 0 to enable a request timeout
	endpoint string
}

// NewClient returns a new API client
func NewClient() *Client {
	// ref:
	// - https://developers.zoom.us/docs/internal-apps/#use-account-credentials-to-get-an-access-token
	// - https://developers.zoom.us/docs/internal-apps/s2s-oauth/#use-account-credentials-to-get-an-access-token
	params := url.Values{}
	params.Add("grant_type", "account_credentials")
	params.Add("account_id", os.Getenv("ZOOM_ACCOUNT_ID"))

	cfg := clientcredentials.Config{
		ClientID:       os.Getenv("ZOOM_CLIENT_ID"),
		ClientSecret:   os.Getenv("ZOOM_CLIENT_SECRET"),
		TokenURL:       "https://zoom.us/oauth/token",
		Scopes:         nil,
		EndpointParams: params,
		AuthStyle:      oauth2.AuthStyleInHeader,
	}

	var uri = url.URL{
		Scheme: "https",
		Host:   apiURI,
		Path:   apiVersion,
	}

	return &Client{
		cfg:      cfg,
		endpoint: uri.String(),
	}
}

type requestV2Opts struct {
	Client         *Client
	Method         HTTPMethod
	URLParameters  interface{}
	Path           string
	DataParameters interface{}
	Ret            interface{}
	// HeadResponse represents responses that don't have a body
	HeadResponse bool
}

func initializeDefault(c *Client) *Client {
	if c == nil {
		if defaultClient == nil {
			defaultClient = NewClient()
		}

		return defaultClient
	}

	return c
}

func (c *Client) executeRequest(opts requestV2Opts) (*http.Response, error) {
	client := c.httpClient()
	req, err := c.httpRequest(opts)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return client.Do(req)
}

func (c *Client) httpRequest(opts requestV2Opts) (*http.Request, error) {
	var buf bytes.Buffer

	// encode body parameters if any
	if err := json.NewEncoder(&buf).Encode(&opts.DataParameters); err != nil {
		return nil, err
	}

	// set URL parameters
	values, err := query.Values(opts.URLParameters)
	if err != nil {
		return nil, err
	}

	// set request URL
	requestURL := c.endpoint + opts.Path
	if len(values) > 0 {
		requestURL += "?" + values.Encode()
	}

	if Debug {
		log.Printf("Request URL: %s", requestURL)
		log.Printf("URL Parameters: %s", values.Encode())
		log.Printf("Body Parameters: %s", buf.String())
	}

	// create HTTP request
	return http.NewRequest(string(opts.Method), requestURL, &buf)
}

func (c *Client) httpClient() *http.Client {
	client := c.cfg.Client(context.TODO())
	if c.Timeout > 0 {
		client.Timeout = c.Timeout
	}
	return client
}

func (c *Client) requestV2(opts requestV2Opts) error {
	// make sure the defaultClient is not nil if we are using it
	c = initializeDefault(c)

	// execute HTTP request
	resp, err := c.executeRequest(opts)
	if err != nil {
		return err
	}

	// If there is no body in response
	if opts.HeadResponse {
		return c.requestV2HeadOnly(resp)
	}

	return c.requestV2WithBody(opts, resp)
}

func (c *Client) requestV2WithBody(opts requestV2Opts, resp *http.Response) error {
	// read HTTP response
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if Debug {
		log.Printf("Response Body: %s", string(body))
	}

	// check for Zoom errors in the response
	if err := checkError(body); err != nil {
		return err
	}

	// unmarshal the response body into the return object
	return json.Unmarshal(body, &opts.Ret)
}

func (c *Client) requestV2HeadOnly(resp *http.Response) error {
	if resp.StatusCode != 204 {
		return errors.New(resp.Status)
	}

	// there were no errors, just return
	return nil
}
