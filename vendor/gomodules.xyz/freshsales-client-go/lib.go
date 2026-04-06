package freshsalesclient

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/gobuffalo/flect"
	"gomodules.xyz/sets"
)

type Client struct {
	client *resty.Client
}

func DefaultFromEnv() *Client {
	return New("https://"+os.Getenv("CRM_BUNDLE_ALIAS"), os.Getenv("CRM_API_TOKEN"))
}

func New(baseURL, token string) *Client {
	return &Client{
		client: resty.New().
			EnableTrace().
			SetBaseURL(baseURL).
			SetHeader("Accept", "application/json").
			SetHeader("Authorization", fmt.Sprintf("Token token=%s", token)),
	}
}

type EntityType string

const (
	EntityContact      EntityType = "Contact"
	EntitySalesAccount EntityType = "SalesAccount"
	EntityDeal         EntityType = "Deal"
)

func (c *Client) CreateContact(contact *Contact) (*Contact, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Contact: contact}).
		SetResult(&APIObject{}).
		Post("/api/contacts")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*APIObject).Contact, nil
}

// Get Contact by id

// ref: https://developers.freshworks.com/crm/api/#view_a_contact
// https://appscode.freshsales.io/contacts/5022967942
//  /api/contacts/[id]
/*
	curl -H "Authorization: Token token=sfg999666t673t7t82" -H "Content-Type: application/json" -X GET "https://domain.freshsales.io/api/contacts/1"
*/
func (c *Client) GetContact(id int) (*Contact, error) {
	resp, err := c.client.R().
		SetResult(APIObject{}).
		Get(fmt.Sprintf("/api/contacts/%d", id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*APIObject).Contact, nil
}

func (c *Client) UpdateContact(contact *Contact) (*Contact, error) {
	resp, err := c.client.R().
		SetBody(APIObject{Contact: contact}).
		SetResult(&APIObject{}).
		Put(fmt.Sprintf("/api/contacts/%d", contact.ID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*APIObject).Contact, nil
}

func (c *Client) DeleteContact(id int) error {
	resp, err := c.client.R().
		Delete(fmt.Sprintf("/api/contacts/%d", id))
	if err != nil {
		return err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) UpsertContact(uniqueIdentifier map[string]string, contact *Contact) (*Contact, error) {
	type UpsertRequest struct {
		UniqueIdentifier map[string]string `json:"unique_identifier"`
		Contact          *Contact          `json:"contact"`
	}
	resp, err := c.client.R().
		SetBody(UpsertRequest{
			UniqueIdentifier: uniqueIdentifier,
			Contact:          contact,
		}).
		SetResult(&APIObject{}).
		Post("/api/contacts/upsert")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*APIObject).Contact, nil
}

func (c *Client) GetContactFilters() ([]ContactView, error) {
	resp, err := c.client.R().
		SetResult(ContactFilters{}).
		Get("/api/contacts/filters")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*ContactFilters).Filters, nil
}

func (c *Client) ListAllContacts() ([]Contact, error) {
	views, err := c.GetContactFilters()
	if err != nil {
		return nil, err
	}
	viewId := -1
	for _, view := range views {
		if view.Name == "All Contacts" {
			viewId = view.ID
			break
		}
	}
	if viewId == -1 {
		return nil, fmt.Errorf("failed to detect view_id for \"All Contacts\"")
	}

	page := 1
	var contacts []Contact
	for {
		resp, err := c.getContactPage(viewId, page)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, resp.Contacts...)
		if page == resp.Meta.TotalPages {
			break
		}
		page++
	}
	return contacts, nil
}

func (c *Client) getContactPage(viewId, page int) (*ListResponse, error) {
	resp, err := c.client.R().
		SetResult(ListResponse{}).
		SetQueryParam("page", strconv.Itoa(page)).
		Get(fmt.Sprintf("/api/contacts/view/%d", viewId))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*ListResponse), nil
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
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
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
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
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
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*LookupResult), nil
}

func (c *Client) CreateAccount(account *SalesAccount) (*SalesAccount, error) {
	resp, err := c.client.R().
		SetBody(AccountObject{SalesAccount: account}).
		SetResult(&AccountObject{}).
		Post("/api/sales_accounts")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*AccountObject).SalesAccount, nil
}

func (c *Client) GetAccount(id int) (*SalesAccount, error) {
	resp, err := c.client.R().
		SetResult(AccountObject{}).
		Get(fmt.Sprintf("/api/sales_accounts/%d", id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*AccountObject).SalesAccount, nil
}

func (c *Client) UpdateAccount(account *SalesAccount) (*SalesAccount, error) {
	resp, err := c.client.R().
		SetBody(AccountObject{SalesAccount: account}).
		SetResult(&AccountObject{}).
		Put(fmt.Sprintf("/api/sales_accounts/%d", account.ID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*AccountObject).SalesAccount, nil
}

func (c *Client) DeleteAccount(id int) error {
	resp, err := c.client.R().
		Delete(fmt.Sprintf("/api/sales_accounts/%d", id))
	if err != nil {
		return err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) GetAccountFilters() ([]AccountView, error) {
	resp, err := c.client.R().
		SetResult(AccountFilters{}).
		Get("/api/sales_accounts/filters")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*AccountFilters).Filters, nil
}

func (c *Client) ListAllAccounts() ([]SalesAccount, error) {
	views, err := c.GetAccountFilters()
	if err != nil {
		return nil, err
	}
	viewId := -1
	for _, view := range views {
		if view.Name == "All Accounts" {
			viewId = view.ID
			break
		}
	}
	if viewId == -1 {
		return nil, fmt.Errorf("failed to detect view_id for \"All Accounts\"")
	}

	page := 1
	var accounts []SalesAccount
	for {
		resp, err := c.getAccountPage(viewId, page)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, resp.SalesAccounts...)
		if page == resp.Meta.TotalPages {
			break
		}
		page++
	}
	return accounts, nil
}

func (c *Client) getAccountPage(viewId, page int) (*AccountListResponse, error) {
	resp, err := c.client.R().
		SetResult(AccountListResponse{}).
		SetQueryParam("page", strconv.Itoa(page)).
		Get(fmt.Sprintf("/api/sales_accounts/view/%d", viewId))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*AccountListResponse), nil
}

func (c *Client) CreateDeal(deal *Deal) (*Deal, error) {
	resp, err := c.client.R().
		SetBody(DealObject{Deal: deal}).
		SetResult(&DealObject{}).
		Post("/api/deals")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*DealObject).Deal, nil
}

func (c *Client) GetDeal(id int) (*Deal, error) {
	resp, err := c.client.R().
		SetResult(DealObject{}).
		Get(fmt.Sprintf("/api/deals/%d", id))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*DealObject).Deal, nil
}

func (c *Client) UpdateDeal(deal *Deal) (*Deal, error) {
	resp, err := c.client.R().
		SetBody(DealObject{Deal: deal}).
		SetResult(&DealObject{}).
		Put(fmt.Sprintf("/api/deals/%d", deal.ID))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*DealObject).Deal, nil
}

func (c *Client) DeleteDeal(id int) error {
	resp, err := c.client.R().
		Delete(fmt.Sprintf("/api/deals/%d", id))
	if err != nil {
		return err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return nil
}

func (c *Client) GetDealFilters() ([]DealView, error) {
	resp, err := c.client.R().
		SetResult(DealFilters{}).
		Get("/api/deals/filters")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*DealFilters).Filters, nil
}

func (c *Client) ListAllDeals() ([]Deal, error) {
	views, err := c.GetDealFilters()
	if err != nil {
		return nil, err
	}
	viewId := -1
	for _, view := range views {
		if view.Name == "All Deals" {
			viewId = view.ID
			break
		}
	}
	if viewId == -1 {
		return nil, fmt.Errorf("failed to detect view_id for \"All Deals\"")
	}

	page := 1
	var deals []Deal
	for {
		resp, err := c.getDealPage(viewId, page)
		if err != nil {
			return nil, err
		}
		deals = append(deals, resp.Deals...)
		if page == resp.Meta.TotalPages {
			break
		}
		page++
	}
	return deals, nil
}

func (c *Client) getDealPage(viewId, page int) (*DealListResponse, error) {
	resp, err := c.client.R().
		SetResult(DealListResponse{}).
		SetQueryParam("page", strconv.Itoa(page)).
		Get(fmt.Sprintf("/api/deals/view/%d", viewId))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() < http.StatusOK || resp.StatusCode() >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("request failed with status code = %d", resp.StatusCode())
	}
	return resp.Result().(*DealListResponse), nil
}
