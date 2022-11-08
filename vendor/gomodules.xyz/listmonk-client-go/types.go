package listmonkclient

import "time"

const (
	ListmonkProd    = "https://listmonk.appscode.com"
	ListmonkTesting = "https://listmonk-testing.appscode.com"

	MailingList_Console   = "06a84456-bfdf-4edf-97c1-7e7d4ad48f67"
	MailingList_KubeDB    = "a5f00cb2-f398-4408-a13a-28b6db8a32ba"
	MailingList_Kubeform  = "cd797afa-04d4-45c8-86e0-642a59b2d7f4"
	MailingList_KubeVault = "b0a46c28-43c3-4048-8059-c3897474b577"
	MailingList_Stash     = "3ab3161e-d02c-42cf-ad96-bb406620d693"
	MailingList_Voyager   = "6c6d1338-bb38-40f6-bab4-ff09c2f6e184"
)

type ListType string

const (
	ListTypePrivate ListType = "private"
	ListTypePublic  ListType = "public"
)

type OptinMode string

const (
	OptinModeSingle OptinMode = "single"
	OptinModeDouble OptinMode = "double"
)

type SubscribeRequest struct {
	Email        string
	Name         string
	MailingLists []string
}

type MailingListRequest struct {
	Name  string    `json:"name"`
	Type  ListType  `json:"type"`
	Optin OptinMode `json:"optin"`
	Tags  []string  `json:"tags,omitempty"`
}

type MailingList struct {
	ID              int       `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	UUID            string    `json:"uuid"`
	Name            string    `json:"name"`
	Type            ListType  `json:"type"`
	Optin           OptinMode `json:"optin"`
	Tags            []string  `json:"tags,omitempty"`
	SubscriberCount int       `json:"subscriber_count"`
}

type ListMailingListResponsePage struct {
	Results []MailingList `json:"results"`
	Total   int           `json:"total"`
	PerPage int           `json:"per_page"`
	Page    int           `json:"page"`
}

type ListMailingListResponse struct {
	Data ListMailingListResponsePage `json:"data"`
}

type GetMailingListResponse struct {
	Data MailingList `json:"data"`
}
