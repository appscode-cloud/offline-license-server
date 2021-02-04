package freshsalesclient

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gobuffalo/flect"
	"gomodules.xyz/sets"
)

type Client struct {
	client *resty.Client
}

func New(host, token string) *Client {
	return &Client{
		client: resty.New().
			EnableTrace().
			SetHostURL(host).
			SetHeader("Accept", "application/json").
			SetHeader("Authorization", fmt.Sprintf("Token token=%s", token)),
	}
}

type EntityType string

const (
	EntityLead         EntityType = "Lead"
	EntityContact      EntityType = "Contact"
	EntitySalesAccount EntityType = "SalesAccount"
	EntityDeal         EntityType = "Deal"
)

func (c *Client) CreateLead(lead *Lead) (*Lead, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Lead: lead}).
		SetResult(&APIObject{}).
		Post("/api/leads")
	if err != nil {
		return nil, err
	}
	return resp.Result().(*APIObject).Lead, nil
}

// Get Lead by id

// ref: https://developer.freshsales.io/api/#view_a_lead
// https://appscode.freshsales.io/leads/5022967942
//  /api/leads/[id]
/*
	curl -H "Authorization: Token token=sfg999666t673t7t82" -H "Content-Type: application/json" -X GET "https://domain.freshsales.io/api/leads/1"
*/
func (c *Client) GetLead(id int) (*Lead, error) {
	resp, err := c.client.R().
		SetResult(APIObject{}).
		Get(fmt.Sprintf("/api/leads/%d", id))
	if err != nil {
		return nil, err
	}
	return resp.Result().(*APIObject).Lead, nil
}

func (c *Client) UpdateLead(lead *Lead) (*Lead, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Lead: lead}).
		SetResult(&APIObject{}).
		Put(fmt.Sprintf("/api/leads/%d", lead.ID))
	if err != nil {
		panic(err)
	}
	return resp.Result().(*APIObject).Lead, nil
}

func (c *Client) UpdateContact(contact *Contact) (*Contact, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Contact: contact}).
		SetResult(&APIObject{}).
		Put(fmt.Sprintf("/api/contacts/%d", contact.ID))
	if err != nil {
		panic(err)
	}
	return resp.Result().(*APIObject).Contact, nil
}

func (c *Client) AddNote(id int64, et EntityType, desc string) (*Note, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Note: &Note{
			Description:    desc,
			TargetableType: string(et),
			TargetableID:   id,
		}}).
		SetResult(&APIObject{}).
		Post("/api/notes")
	if err != nil {
		return nil, err
	}
	return resp.Result().(*APIObject).Note, nil
}

func (c *Client) Search(str string, et EntityType, more ...EntityType) ([]Entity, error) {
	entities := sets.NewString()
	for _, e := range append(more, et) {
		entities.Insert(strings.ToLower(flect.Underscore(string(e))))
	}

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"q":       str,
			"include": strings.Join(entities.List(), ","),
		}).
		SetResult(SearchResults{}).
		Get("/api/search")
	if err != nil {
		return nil, err
	}
	return *(resp.Result().(*SearchResults)), nil
}

func (c *Client) LookupByEmail(email string, et EntityType, more ...EntityType) (*LookupResult, error) {
	entities := sets.NewString()
	for _, e := range append(more, et) {
		entities.Insert(strings.ToLower(flect.Underscore(string(e))))
	}

	resp, err := c.client.R().
		SetQueryParams(map[string]string{
			"q":        email,
			"f":        "email",
			"entities": strings.Join(entities.List(), ","),
		}).
		SetResult(LookupResult{}).
		Get("/api/lookup")
	if err != nil {
		return nil, err
	}
	return resp.Result().(*LookupResult), nil
}
