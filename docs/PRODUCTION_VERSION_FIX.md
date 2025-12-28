# Version Check Production Fix

## Problem

Version check failed silently in production due to:

- No proxy support (corporate firewalls block api.github.com)
- No TLS certificate handling (corporate proxies with custom CAs)
- GitHub rate limiting (60/hour unauthenticated)
- 3-second timeout insufficient for production networks
- Silent error suppression

## Solution

Production-ready HTTP client with:

- `http.ProxyFromEnvironment` - respects standard proxy vars
- TLS 1.2+ with optional InsecureSkipVerify for corporate CAs
- GitHub token auth (5000/hour rate limit)
- 10-second timeout
- 3 retry attempts with exponential backoff
- Debug mode for troubleshooting

## Files Modified

- `ztictl/pkg/version/checker.go` - Added createHTTPClient() with production features
- `docs/VERSION_CHECKING.md` - Environment variable reference

## Environment Variables

- `ZTICTL_SKIP_VERSION_CHECK` - disable check
- `ZTICTL_DEBUG` - show errors
- `ZTICTL_INSECURE_SKIP_VERIFY` - skip TLS verification (corporate CAs only)
- `ZTICTL_HTTPS_PROXY` - custom proxy override
- `GITHUB_TOKEN` - authentication for rate limits
