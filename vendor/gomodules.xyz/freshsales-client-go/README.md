# freshsales-client-go

Go client for Freshsales CRM API.

## Installation

```bash
go get gomodules.xyz/freshsales-client-go
```

## Usage

```go
package main

import (
    "fmt"
    "gomodules.xyz/freshsales-client-go"
)

func main() {
    client := freshsalesclient.DefaultFromEnv()
    // or
    // client := freshsalesclient.New("https://domain.freshsales.io", "api-token")
    
    // List all contacts
    contacts, err := client.ListAllContacts()
    
    // Create a contact
    contact, err := client.CreateContact(&freshsalesclient.Contact{
        FirstName: "John",
        LastName:  "Doe",
        Email:     "john@example.com",
    })
    
    // Get a contact
    contact, err := client.GetContact(123)
    
    // Update a contact
    contact, err := client.UpdateContact(&freshsalesclient.Contact{
        ID:        123,
        FirstName: "Jane",
    })
    
    // Delete a contact
    err := client.DeleteContact(123)
    
    // Upsert a contact
    contact, err := client.UpsertContact(
        map[string]string{"emails": "john@example.com"},
        &freshsalesclient.Contact{FirstName: "John"},
    )
    
    // Accounts
    accounts, err := client.ListAllAccounts()
    account, err := client.CreateAccount(&freshsalesclient.SalesAccount{Name: "Acme"})
    account, err := client.GetAccount(123)
    account, err := client.UpdateAccount(account)
    err := client.DeleteAccount(123)
    
    // Deals
    deals, err := client.ListAllDeals()
    deal, err := client.CreateDeal(&freshsalesclient.Deal{
        Name:           "Big Deal",
        Amount:         50000,
        SalesAccountID: 1,
    })
    deal, err := client.GetDeal(123)
    deal, err := client.UpdateDeal(deal)
    err := client.DeleteDeal(123)
    
    // Notes
    note, err := client.AddNote(123, freshsalesclient.EntityDeal, "Follow up next week")
    
    // Search
    results, err := client.Search("john", freshsalesclient.EntityContact)
    
    // Lookup by email
    result, err := client.LookupByEmail("john@example.com", freshsalesclient.EntityContact)
}
```

## Environment Variables

- `CRM_BUNDLE_ALIAS` - Freshsales domain (e.g., `domain.freshsales.io`)
- `CRM_API_TOKEN` - API authentication token