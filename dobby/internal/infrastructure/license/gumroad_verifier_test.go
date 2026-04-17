package infralicense_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	infralicense "github.com/dobby/filemanager/internal/infrastructure/license"
	"github.com/dobby/filemanager/internal/domain/license"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func createVerifierWithServer(handler http.HandlerFunc) (*infralicense.GumroadVerifier, *httptest.Server) {
	srv := httptest.NewServer(handler)
	return infralicense.NewGumroadVerifier(srv.URL), srv
}

func givenGumroadResponds(body string, status int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GumroadVerifier.Verify — response parsing
// ─────────────────────────────────────────────────────────────────────────────

// S-2-1
func TestGumroadVerifier_Verify_WhenSuccessFalse_ReturnsErrInvalidLicenseKey(t *testing.T) {
	// Arrange
	verifier, srv := createVerifierWithServer(
		givenGumroadResponds(`{"success":false}`, http.StatusOK),
	)
	defer srv.Close()

	// Act
	err := verifier.Verify(context.Background(), "any-key")

	// Assert
	assert.ErrorIs(t, err, license.ErrInvalidLicenseKey)
}

// S-3-1
func TestGumroadVerifier_Verify_WhenUsesGreaterThanOne_ReturnsErrLicenseAlreadyUsed(t *testing.T) {
	// Arrange
	verifier, srv := createVerifierWithServer(
		givenGumroadResponds(`{"success":true,"uses":2}`, http.StatusOK),
	)
	defer srv.Close()

	// Act
	err := verifier.Verify(context.Background(), "any-key")

	// Assert
	assert.ErrorIs(t, err, license.ErrLicenseAlreadyUsed)
}

// S-1-1
func TestGumroadVerifier_Verify_WhenSuccessTrueAndUsesOne_ReturnsNil(t *testing.T) {
	// Arrange
	verifier, srv := createVerifierWithServer(
		givenGumroadResponds(`{"success":true,"uses":1}`, http.StatusOK),
	)
	defer srv.Close()

	// Act
	err := verifier.Verify(context.Background(), "any-key")

	// Assert
	assert.NoError(t, err)
}
