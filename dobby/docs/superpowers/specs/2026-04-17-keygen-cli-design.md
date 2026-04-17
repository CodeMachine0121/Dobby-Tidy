# Keygen CLI Tool Design

**Date:** 2026-04-17  
**Status:** Approved

## Overview

A standalone CLI tool that generates batches of valid DOBBY license keys for manual upload to Gumroad's Unique Codes feature. When a customer purchases the product, Gumroad automatically delivers one key via the receipt email.

## Context

- The Dobby app validates license keys **offline** using HMAC-SHA256 with a shared secret embedded in the binary.
- Key format: `DOBBY-XXXX-XXXX-XXXX` where the last segment is a 4-char HMAC check derived from the first two segments.
- Key generation logic already exists in `internal/domain/license/license_key_validator.go` (`GenerateKey(p1, p2 string) string`).

## Delivery Flow

1. Developer runs `go run ./cmd/keygen -n 10000 -o keys.txt`
2. Developer uploads `keys.txt` to Gumroad product → Unique Codes
3. Customer purchases → Gumroad emails receipt with one key attached
4. Customer enters key in the Dobby app → offline HMAC validation passes → activated

## Implementation

**New file:** `cmd/keygen/main.go` (no changes to existing code)

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-n` | int | 10000 | Number of keys to generate |
| `-o` | string | stdout | Output file path |

**Algorithm:**

1. Use `crypto/rand` to generate random p1 and p2 (4 chars each, charset `A-Z0-9`)
2. Call `license.NewLicenseKeyValidator().GenerateKey(p1, p2)`
3. Track generated keys in a `map[string]struct{}` to deduplicate within the batch
4. Write one key per line to the output destination

**Example output:**
```
DOBBY-A3K9-ZX12-F7QR
DOBBY-BT4M-WE83-P2NV
...
```

## Out of Scope

- No CI/CD integration (manual use only for now)
- No persistent record of previously generated keys across runs
- No server-side key validation
