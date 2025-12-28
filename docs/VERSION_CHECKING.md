# Version Checking

## Architecture

Version checker uses production-ready HTTP client to check GitHub releases API.

### Design Decisions

**Proxy Support**

- `http.ProxyFromEnvironment` respects standard HTTP_PROXY, HTTPS_PROXY, NO_PROXY
- Custom override via ZTICTL_HTTPS_PROXY for tool-specific proxy configuration
- Rationale: Corporate environments require proxy support without modifying system-wide settings

**TLS Configuration**

- Minimum TLS 1.2
- Optional InsecureSkipVerify for corporate proxy with custom certificates
- Rationale: Balance security with corporate environment compatibility

**Retry Logic**

- 3 attempts with exponential backoff (1s base delay)
- Skip retry on 4xx client errors
- Rationale: Transient network failures common in production; client errors are permanent

**Rate Limiting**

- Unauthenticated: 60 requests/hour (GitHub limit)
- With GITHUB_TOKEN: 5000 requests/hour
- 24-hour cache to minimize API calls
- Rationale: Prevent rate limit exhaustion in CI/CD environments

**Error Handling**

- Silent by default (non-blocking)
- Debug mode (ZTICTL_DEBUG) shows detailed errors
- Disable option (ZTICTL_SKIP_VERSION_CHECK) for air-gapped environments
- Rationale: Version check should never block primary operations

## Environment Variables

| Variable                    | Purpose                   |
| --------------------------- | ------------------------- |
| ZTICTL_SKIP_VERSION_CHECK   | Disable version checking  |
| ZTICTL_DEBUG                | Show detailed errors      |
| ZTICTL_VERBOSE_VERSION      | Show success messages     |
| ZTICTL_HTTPS_PROXY          | Custom proxy override     |
| ZTICTL_INSECURE_SKIP_VERIFY | Skip TLS verification     |
| GITHUB_TOKEN                | GitHub API authentication |

## Implementation

- File: `ztictl/pkg/version/checker.go`
- Function: `createHTTPClient()` - production HTTP client factory
- Function: `CheckLatestVersion()` - retry logic and caching
- Function: `PrintVersionWithCheck()` - output formatting
