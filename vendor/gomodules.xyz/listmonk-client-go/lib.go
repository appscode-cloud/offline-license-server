package listmonkclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func New(host, username, password string) *Client {
	return &Client{
		client: resty.New().
			EnableTrace().
			SetHostURL(host).
			SetHeader("Accept", "application/json").
			SetHeader("Content-Type", "application/json;charset=UTF-8").
			SetBasicAuth(username, password),
	}
}

func (c *Client) SubscribeToList(req SubscribeRequest) error {
	if len(req.MailingLists) == 0 {
		return fmt.Errorf("missing list id: %+v", req)
	}

	params := url.Values{}
	params.Add("email", req.Email)
	params.Add("name", req.Name)
	for _, listID := range req.MailingLists {
		params.Add("l", listID)
	}

	resp, err := c.client.R().
		SetFormDataFromValues(params).
		SetHeader("Authorization", "").
		Post("/subscription/form")
	if err != nil {
		return err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) CreateListIfMissing(req MailingListRequest) (*MailingList, error) {
	lists, err := c.GetAllLists()
	if err != nil {
		return nil, err
	}
	for _, l := range lists {
		if l.Name == req.Name {
			return &l, nil
		}
	}
	return c.CreateList(req)
}

func (c *Client) CreateList(req MailingListRequest) (*MailingList, error) {
	resp, err := c.client.R().
		SetBody(req).
		SetResult(&GetMailingListResponse{}).
		Post("/api/lists")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return &(resp.Result().(*GetMailingListResponse).Data), nil
}

func (c *Client) GetAllLists() ([]MailingList, error) {
	var out []MailingList

	page := 1
	for {
		resp, err := c.getListPage(page)
		if err != nil {
			return nil, err
		}
		out = append(out, resp.Results...)
		if resp.Total > len(out) {
			page++
		} else {
			break
		}
	}
	return out, nil
}

func (c *Client) getListPage(page int) (*ListMailingListResponsePage, error) {
	resp, err := c.client.R().
		SetResult(ListMailingListResponse{}).
		SetQueryParams(map[string]string{
			"page":     strconv.Itoa(page),
			"order_by": "created_at",
			"order":    "asc",
		}).
		Get("/api/lists")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return &(resp.Result().(*ListMailingListResponse).Data), nil
}
