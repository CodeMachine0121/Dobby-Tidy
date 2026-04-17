package main

import (
	"testing"

	"github.com/dobby/filemanager/internal/domain/license"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createValidator() *license.LicenseKeyValidator {
	return license.NewLicenseKeyValidator()
}

// ─────────────────────────────────────────────────────────────────────────────
// generateKeys
// ─────────────────────────────────────────────────────────────────────────────

// S-2-1
func TestGenerateKeys_ReturnsExactCount(t *testing.T) {
	// Arrange
	n := 50

	// Act
	keys, err := generateKeys(n)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, keys, n)
}

// S-4-1
func TestGenerateKeys_AllUnique(t *testing.T) {
	// Arrange
	n := 1000

	// Act
	keys, err := generateKeys(n)

	// Assert
	assert.NoError(t, err)
	seen := make(map[string]struct{}, n)
	for _, k := range keys {
		seen[k] = struct{}{}
	}
	assert.Len(t, seen, n)
}

// S-5-1
func TestGenerateKeys_EachKeyPassesHMAC(t *testing.T) {
	// Arrange
	validator := createValidator()

	// Act
	keys, err := generateKeys(10)

	// Assert
	assert.NoError(t, err)
	for _, k := range keys {
		assert.NoError(t, validator.Validate(k))
	}
}

// S-6-1
func TestGenerateKeys_ZeroCount_ReturnsEmptySlice(t *testing.T) {
	// Arrange — count = 0

	// Act
	keys, err := generateKeys(0)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, keys)
}

// S-7-1
func TestGenerateKeys_NegativeCount_ReturnsError(t *testing.T) {
	// Arrange — count = -1

	// Act
	_, err := generateKeys(-1)

	// Assert
	assert.Error(t, err)
}
