package freshsalesclient

import "time"

type SearchResults []Entity

type Entity struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar"`
	CompanyName string `json:"company_name,omitempty"`
}

type CompanyType struct {
	Partial  bool   `json:"partial,omitempty"`
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Position int    `json:"position,omitempty"`
}

type Company struct {
	ID                int64       `json:"id,omitempty"`
	Name              string      `json:"name,omitempty"`
	Address           string      `json:"address,omitempty"`
	City              string      `json:"city,omitempty"`
	State             string      `json:"state,omitempty"`
	Zipcode           string      `json:"zipcode,omitempty"`
	Country           string      `json:"country,omitempty"`
	NumberOfEmployees int         `json:"number_of_employees,omitempty"`
	AnnualRevenue     int         `json:"annual_revenue,omitempty"`
	Website           string      `json:"website,omitempty"`
	Phone             string      `json:"phone,omitempty"`
	IndustryTypeID    int64       `json:"industry_type_id,omitempty"`
	IndustryType      CompanyType `json:"industry_type,omitempty"`
	BusinessTypeID    int64       `json:"business_type_id,omitempty"`
	BusinessType      CompanyType `json:"business_type,omitempty"`
}

type Currency struct {
	Partial      bool        `json:"partial,omitempty"`
	ID           int64       `json:"id,omitempty"`
	IsActive     bool        `json:"is_active,omitempty"`
	CurrencyCode string      `json:"currency_code,omitempty"`
	ExchangeRate string      `json:"exchange_rate,omitempty"`
	CurrencyType int         `json:"currency_type,omitempty"`
	ScheduleInfo interface{} `json:"schedule_info,omitempty"`
}

type Deal struct {
	ID                 int64       `json:"id,omitempty"`
	Name               string      `json:"name,omitempty"`
	Amount             float64     `json:"amount,omitempty"`
	CurrencyID         int64       `json:"currency_id,omitempty"`
	BaseCurrencyAmount float64     `json:"base_currency_amount,omitempty"`
	ExpectedClose      *time.Time  `json:"expected_close,omitempty"`
	DealProductID      int64       `json:"deal_product_id,omitempty"`
	DealProduct        interface{} `json:"deal_product,omitempty"`
	Currency           Currency    `json:"currency,omitempty"`
	ProductID          int         `json:"product_id,omitempty"`
}

type Links struct {
	Conversations        string `json:"conversations,omitempty"`
	TimelineFeeds        string `json:"timeline_feeds,omitempty"`
	DocumentAssociations string `json:"document_associations,omitempty"`
	Notes                string `json:"notes,omitempty"`
	Tasks                string `json:"tasks,omitempty"`
	Appointments         string `json:"appointments,omitempty"`
	Reminders            string `json:"reminders,omitempty"`
	Duplicates           string `json:"duplicates,omitempty"`
	Connections          string `json:"connections,omitempty"`
}

type EmailInfo struct {
	ID        int64       `json:"id,omitempty"`
	Value     string      `json:"value,omitempty"`
	IsPrimary bool        `json:"is_primary,omitempty"`
	Label     interface{} `json:"label,omitempty"`
	Destroy   bool        `json:"_destroy,omitempty"`
}

type CustomFields struct {
	Interest              string `json:"cf_interest,omitempty"`
	Github                string `json:"cf_github,omitempty"`
	KubernetesSetup       string `json:"cf_kubernetes_setup,omitempty"`
	CalendlyMeetingAgenda string `json:"cf_calendly_meeting_agenda,omitempty"`
}

type Contact struct {
	ID                             int64        `json:"id"`
	FirstName                      string       `json:"first_name"`
	LastName                       string       `json:"last_name"`
	DisplayName                    string       `json:"display_name"`
	Avatar                         string       `json:"avatar"`
	JobTitle                       string       `json:"job_title"`
	City                           string       `json:"city"`
	State                          string       `json:"state"`
	Zipcode                        string       `json:"zipcode"`
	Country                        string       `json:"country"`
	Email                          string       `json:"email"`
	Emails                         []EmailInfo  `json:"emails"`
	TimeZone                       string       `json:"time_zone"`
	WorkNumber                     string       `json:"work_number"`
	MobileNumber                   string       `json:"mobile_number"`
	Address                        string       `json:"address"`
	LastSeen                       string       `json:"last_seen"`
	LeadScore                      int          `json:"lead_score"`
	LastContacted                  time.Time    `json:"last_contacted"`
	OpenDealsAmount                string       `json:"open_deals_amount"`
	WonDealsAmount                 string       `json:"won_deals_amount"`
	Links                          Links        `json:"links"`
	LastContactedSalesActivityMode string       `json:"last_contacted_sales_activity_mode"`
	CustomField                    CustomFields `json:"custom_field"`
	CreatedAt                      time.Time    `json:"created_at"`
	UpdatedAt                      time.Time    `json:"updated_at"`
	Keyword                        string       `json:"keyword"`
	Medium                         string       `json:"medium"`
	LastContactedMode              string       `json:"last_contacted_mode"`
	RecentNote                     string       `json:"recent_note"`
	WonDealsCount                  int          `json:"won_deals_count"`
	LastContactedViaSalesActivity  time.Time    `json:"last_contacted_via_sales_activity"`
	CompletedSalesSequences        string       `json:"completed_sales_sequences"`
	ActiveSalesSequences           string       `json:"active_sales_sequences"`
	WebFormIds                     string       `json:"web_form_ids"`
	OpenDealsCount                 int          `json:"open_deals_count"`
	LastAssignedAt                 *time.Time   `json:"last_assigned_at"`
	Facebook                       string       `json:"facebook"`
	Twitter                        string       `json:"twitter"`
	Linkedin                       string       `json:"linkedin"`
	IsDeleted                      bool         `json:"is_deleted"`
	TeamUserIds                    string       `json:"team_user_ids"`
	ExternalId                     string       `json:"external_id"`
	WorkEmail                      string       `json:"work_email"`
	SubscriptionStatus             int          `json:"subscription_status"`
	SubscriptionTypes              string       `json:"subscription_types"`
	UnsubscriptionReason           string       `json:"unsubscription_reason"`
	OtherUnsubscriptionReason      string       `json:"other_unsubscription_reason"`
	CustomerFit                    int          `json:"customer_fit"`
	WhatsappSubscriptionStatus     int          `json:"whatsapp_subscription_status"`
	SmsSubscriptionStatus          int          `json:"sms_subscription_status"`
	LastSeenChat                   string       `json:"last_seen_chat"`
	FirstSeenChat                  string       `json:"first_seen_chat"`
	Locale                         string       `json:"locale"`
	TotalSessions                  string       `json:"total_sessions"`
	SystemTags                     []string     `json:"system_tags"`
	FirstCampaign                  string       `json:"first_campaign"`
	FirstMedium                    string       `json:"first_medium"`
	FirstSource                    string       `json:"first_source"`
	LastCampaign                   string       `json:"last_campaign"`
	LastMedium                     string       `json:"last_medium"`
	LastSource                     string       `json:"last_source"`
	LatestCampaign                 string       `json:"latest_campaign"`
	LatestMedium                   string       `json:"latest_medium"`
	LatestSource                   string       `json:"latest_source"`
	McrId                          int64        `json:"mcr_id"`
	PhoneNumbers                   []string     `json:"phone_numbers"`
	Tags                           []string     `json:"tags"`
}

type LookupResult struct {
	Contacts struct {
		Contacts []Contact `json:"contacts,omitempty"`
	} `json:"contacts,omitempty"`
}

type Note struct {
	Description    string `json:"description,omitempty"`
	TargetableType string `json:"targetable_type,omitempty"`
	TargetableID   int64  `json:"targetable_id,omitempty"`

	ID        int64      `json:"id,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type APIObject struct {
	Contact *Contact `json:"contact,omitempty"`
	Note    *Note    `json:"note,omitempty"`
}

type BaseNoteDescription struct {
	Event  string     `json:"event,omitempty"`
	Client ClientInfo `json:"client,omitempty"`
}

type ClientInfo struct {
	OS       string      `json:"os,omitempty"`
	Device   string      `json:"device,omitempty"`
	Location GeoLocation `json:"location,omitempty"`
}

type GeoLocation struct {
	IP          string `json:"ip,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	City        string `json:"city,omitempty"`
	Country     string `json:"country,omitempty"`
	Coordinates string `json:"coordinates,omitempty"`
}

type ContactView struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ModelClassName string `json:"model_class_name"`
	UserID         int    `json:"user_id"`
	IsDefault      bool   `json:"is_default"`
	IsPublic       bool   `json:"is_public"`
	UpdatedAt      string `json:"updated_at"`
}

type ContactFilters struct {
	Filters []ContactView `json:"filters"`
}

type ListMeta struct {
	TotalPages int `json:"total_pages"`
	Total      int `json:"total"`
}

type ListResponse struct {
	Contacts []Contact `json:"contacts"`
	Meta     ListMeta  `json:"meta"`
}
