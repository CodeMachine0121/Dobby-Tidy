# Gumroad License Verification Design

**Date:** 2026-04-17  
**Status:** Approved

## Overview

Replace the existing offline HMAC license key validation (`DOBBY-XXXX-XXXX-XXXX` format) with Gumroad's native license key system. Activation requires a one-time internet connection; after activation the app operates fully offline.

## Context

- **Product permalink:** `dobby-tidy` (`afternoonjames.gumroad.com/l/dobby-tidy`)
- **Gumroad verify endpoint:** `POST https://api.gumroad.com/v2/licenses/verify`
- **Machine limit:** 1 machine per key (enforced via `increment_uses_count=true` + checking `uses == 1`)
- **No deactivation/transfer:** users who change machines must contact the developer manually
- **Existing local storage** (`license.dat`) is kept unchanged — it simply stores the Gumroad key string instead of a DOBBY key

## Activation Flow

```
User enters Gumroad key in Settings page
  → App.ActivateLicense(key)
  → LicenseService.ActivateLicense(ctx, key)
  → IGumroadVerifier.Verify(ctx, key)
      POST https://api.gumroad.com/v2/licenses/verify
        product_id=dobby-tidy
        license_key=<key>
        increment_uses_count=true
      Response cases:
        success=false          → ErrInvalidLicenseKey
        success=true, uses > 1 → ErrLicenseAlreadyUsed
        success=true, uses == 1 → OK
  → license.Activate(key, machineID)
  → repo.Save(license)

Subsequent launches:
  → repo.Load() → offline check only, no API call
```

## Components

### New: `IGumroadVerifier` (application layer port)

```go
type IGumroadVerifier interface {
    Verify(ctx context.Context, licenseKey string) error
}
```

Lives in `internal/application/license_service.go` alongside the service.

### New: `GumroadVerifier` (infrastructure implementation)

File: `internal/infrastructure/license/gumroad_verifier.go`

Responsibilities:
- POST to Gumroad API with `product_id`, `license_key`, `increment_uses_count=true`
- Parse JSON response
- Return typed errors for: network failure, invalid key, already-used key

### Modified: `LicenseService.ActivateLicense`

Remove call to `validator.Validate(key)`. Replace with `verifier.Verify(ctx, key)`. Remove `validator` field entirely.

### Removed

| Path | Reason |
|------|--------|
| `internal/domain/license/license_key_validator.go` | HMAC logic no longer needed |
| `cmd/keygen/` | Pre-generated keys no longer used |

### Unchanged

- `internal/domain/license/license.go` — `Activate(key, machineID)` accepts any string
- `internal/infrastructure/license/local_license_repository.go` — stores key as-is
- Frontend Settings page — no changes needed

## Error Handling

| Situation | Error returned to UI |
|-----------|---------------------|
| Network failure | `"無法連線至驗證伺服器，請確認網路連線"` |
| `success: false` from Gumroad | `"無效的 license key"` |
| `uses > 1` | `"此 key 已在另一台機器上使用"` |
| Already activated locally | `ErrAlreadyActivated` (existing) |

## Testing

**`LicenseService` unit tests** (mock `IGumroadVerifier`):
- Verify succeeds → license activated and saved
- Verify returns error → error propagated, nothing saved
- License already activated → `ErrAlreadyActivated` without calling verifier

**`GumroadVerifier` unit tests** (no real HTTP):
- Parse `success=false` response → `ErrInvalidLicenseKey`
- Parse `success=true, uses=2` response → `ErrLicenseAlreadyUsed`
- Parse `success=true, uses=1` response → nil

## Value Objects Changes

Remove from `internal/domain/license/value_objects.go`:
- `ErrInvalidLicenseKeyFormat`
- `ErrInvalidLicenseKey`

Add:
- `ErrInvalidLicenseKey` (reuse name, new meaning: Gumroad rejected the key)
- `ErrLicenseAlreadyUsed` (new: key used on another machine)

## Out of Scope

- Deactivation / machine transfer
- Offline activation fallback
- Key revocation
