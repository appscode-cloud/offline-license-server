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
	Interest              interface{} `json:"cf_interest,omitempty"`
	Github                interface{} `json:"cf_github,omitempty"`
	KubernetesSetup       string      `json:"cf_kubernetes_setup,omitempty"`
	CalendlyMeetingAgenda interface{} `json:"cf_calendly_meeting_agenda,omitempty"`
}

type Lead struct {
	ID                             int64         `json:"id,omitempty"`
	JobTitle                       string        `json:"job_title,omitempty"`
	Department                     string        `json:"department,omitempty"`
	Email                          string        `json:"email,omitempty"`
	Emails                         []EmailInfo   `json:"emails,omitempty"`
	WorkNumber                     string        `json:"work_number,omitempty"`
	MobileNumber                   string        `json:"mobile_number,omitempty"`
	Address                        string        `json:"address,omitempty"`
	City                           string        `json:"city,omitempty"`
	State                          string        `json:"state,omitempty"`
	Zipcode                        string        `json:"zipcode,omitempty"`
	Country                        string        `json:"country,omitempty"`
	TimeZone                       string        `json:"time_zone,omitempty"`
	DoNotDisturb                   bool          `json:"do_not_disturb,omitempty"`
	DisplayName                    string        `json:"display_name,omitempty"`
	Avatar                         string        `json:"avatar,omitempty"`
	Keyword                        string        `json:"keyword,omitempty"`
	Medium                         string        `json:"medium,omitempty"`
	LastSeen                       *time.Time    `json:"last_seen,omitempty"`
	LastContacted                  *time.Time    `json:"last_contacted,omitempty"`
	LeadScore                      int           `json:"lead_score,omitempty"`
	LeadQuality                    string        `json:"lead_quality,omitempty"`
	StageUpdatedTime               *time.Time    `json:"stage_updated_time,omitempty"`
	FirstName                      string        `json:"first_name,omitempty"`
	LastName                       string        `json:"last_name,omitempty"`
	Company                        Company       `json:"company,omitempty"`
	Deal                           Deal          `json:"deal,omitempty"`
	Links                          Links         `json:"links,omitempty"`
	CustomField                    CustomFields  `json:"custom_field,omitempty"`
	CreatedAt                      string        `json:"created_at,omitempty"`
	UpdatedAt                      string        `json:"updated_at,omitempty"`
	LastContactedSalesActivityMode string        `json:"last_contacted_sales_activity_mode,omitempty"`
	HasAuthority                   bool          `json:"has_authority,omitempty"`
	EmailStatus                    string        `json:"email_status,omitempty"`
	LastContactedMode              string        `json:"last_contacted_mode,omitempty"`
	RecentNote                     string        `json:"recent_note,omitempty"`
	LastContactedViaChat           *time.Time    `json:"last_contacted_via_chat,omitempty"`
	LastContactedViaSalesActivity  string        `json:"last_contacted_via_sales_activity,omitempty"`
	CompletedSalesSequences        int           `json:"completed_sales_sequences,omitempty"`
	ActiveSalesSequences           int           `json:"active_sales_sequences,omitempty"`
	WebFormIds                     string        `json:"web_form_ids,omitempty"`
	LastAssignedAt                 *time.Time    `json:"last_assigned_at,omitempty"`
	Tags                           []string      `json:"tags,omitempty"`
	Facebook                       string        `json:"facebook,omitempty"`
	Twitter                        string        `json:"twitter,omitempty"`
	Linkedin                       string        `json:"linkedin,omitempty"`
	IsDeleted                      bool          `json:"is_deleted,omitempty"`
	TeamUserIds                    interface{}   `json:"team_user_ids,omitempty"`
	SubscriptionStatus             int           `json:"subscription_status,omitempty"`
	PhoneNumbers                   []interface{} `json:"phone_numbers,omitempty"`
}

type Contact struct {
	ID                             int64        `json:"id,omitempty"`
	FirstName                      string       `json:"first_name,omitempty"`
	LastName                       string       `json:"last_name,omitempty"`
	DisplayName                    string       `json:"display_name,omitempty"`
	Avatar                         string       `json:"avatar,omitempty"`
	JobTitle                       string       `json:"job_title,omitempty"`
	City                           string       `json:"city,omitempty"`
	State                          string       `json:"state,omitempty"`
	Zipcode                        string       `json:"zipcode,omitempty"`
	Country                        string       `json:"country,omitempty"`
	Email                          string       `json:"email,omitempty"`
	Emails                         []EmailInfo  `json:"emails,omitempty"`
	DoNotDisturb                   bool         `json:"do_not_disturb,omitempty"`
	HasAuthority                   bool         `json:"has_authority,omitempty"`
	TimeZone                       string       `json:"time_zone,omitempty"`
	Department                     string       `json:"department,omitempty"`
	WorkNumber                     string       `json:"work_number,omitempty"`
	MobileNumber                   string       `json:"mobile_number,omitempty"`
	Address                        string       `json:"address,omitempty"`
	LastSeen                       *time.Time   `json:"last_seen,omitempty"`
	LeadScore                      int          `json:"lead_score,omitempty"`
	LeadQuality                    string       `json:"lead_quality,omitempty"`
	LastContacted                  *time.Time   `json:"last_contacted,omitempty"`
	OpenDealsAmount                string       `json:"open_deals_amount,omitempty"`
	WonDealsAmount                 string       `json:"won_deals_amount,omitempty"`
	Links                          Links        `json:"links,omitempty"`
	LastContactedSalesActivityMode string       `json:"last_contacted_sales_activity_mode,omitempty"`
	CustomField                    CustomFields `json:"custom_field,omitempty"`
	CreatedAt                      string       `json:"created_at,omitempty"`
	UpdatedAt                      string       `json:"updated_at,omitempty"`
	Keyword                        string       `json:"keyword,omitempty"`
	Medium                         string       `json:"medium,omitempty"`
	EmailStatus                    string       `json:"email_status,omitempty"`
	LastContactedMode              string       `json:"last_contacted_mode,omitempty"`
	RecentNote                     string       `json:"recent_note,omitempty"`
	LastContactedViaChat           *time.Time   `json:"last_contacted_via_chat,omitempty"`
	WonDealsCount                  int          `json:"won_deals_count,omitempty"`
	LastContactedViaSalesActivity  string       `json:"last_contacted_via_sales_activity,omitempty"`
	CompletedSalesSequences        int          `json:"completed_sales_sequences,omitempty"`
	ActiveSalesSequences           int          `json:"active_sales_sequences,omitempty"`
	WebFormIds                     string       `json:"web_form_ids,omitempty"`
	OpenDealsCount                 int          `json:"open_deals_count,omitempty"`
	LastAssignedAt                 *time.Time   `json:"last_assigned_at,omitempty"`
	Tags                           []string     `json:"tags,omitempty"`
	Facebook                       string       `json:"facebook,omitempty"`
	Twitter                        string       `json:"twitter,omitempty"`
	Linkedin                       string       `json:"linkedin,omitempty"`
	IsDeleted                      bool         `json:"is_deleted,omitempty"`
	TeamUserIds                    interface{}  `json:"team_user_ids,omitempty"`
	SubscriptionStatus             int          `json:"subscription_status,omitempty"`
	CustomerFit                    int          `json:"customer_fit,omitempty"`
}

type LookupResult struct {
	Leads struct {
		Leads []Lead `json:"leads,omitempty"`
	} `json:"leads,omitempty"`
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
	Lead    *Lead    `json:"lead,omitempty"`
	Contact *Contact `json:"contact,omitempty"`
	Note    *Note    `json:"note,omitempty"`
}
