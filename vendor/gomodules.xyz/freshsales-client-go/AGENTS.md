# AGENTS.md

## Setup

Requires environment variables:
- `CRM_BUNDLE_ALIAS` - Freshsales domain (e.g., `domain.freshsales.io`)
- `CRM_API_TOKEN` - API authentication token

Use `DefaultFromEnv()` to create a client, or `New(baseURL, token)` for explicit values.

## Run commands

```bash
go build ./...
go test ./...
```

## Structure

- Root package (`lib.go`, `types.go`) - client library
- `freshsales-export-leads/`, `freshsales-classify-leads/`, `client-demo/` - example programs

## Notes

- No tests exist in this repo
- Uses `resty/v2` for HTTP client
