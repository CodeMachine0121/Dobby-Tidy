package infralicense

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dobby/filemanager/internal/domain/license"
)

const (
	gumroadVerifyURL = "https://api.gumroad.com/v2/licenses/verify"
	gumroadProductID = "80RHP185znin74AyLPo3BA=="
)

type gumroadVerifyResponse struct {
	Success bool `json:"success"`
	Uses    int  `json:"uses"`
}

// GumroadVerifier verifies a license key against the Gumroad API.
type GumroadVerifier struct {
	endpoint string
	client   *http.Client
}

// NewGumroadVerifier creates a verifier targeting the given endpoint (production or test server).
func NewGumroadVerifier(endpoint string) *GumroadVerifier {
	return &GumroadVerifier{endpoint: endpoint, client: &http.Client{}}
}

// NewProductionGumroadVerifier creates a verifier targeting the real Gumroad API.
func NewProductionGumroadVerifier() *GumroadVerifier {
	return NewGumroadVerifier(gumroadVerifyURL)
}

func (v *GumroadVerifier) Verify(ctx context.Context, licenseKey string) error {
	form := url.Values{}
	form.Set("product_id", gumroadProductID)
	form.Set("license_key", strings.TrimSpace(licenseKey))
	form.Set("increment_uses_count", "true")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint,
		strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("無法連線至驗證伺服器，請確認網路連線: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("無法連線至驗證伺服器，請確認網路連線: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("無法連線至驗證伺服器，請確認網路連線: %w", err)
	}
	log.Printf("[gumroad] status=%d body=%s", resp.StatusCode, string(bodyBytes))
	_ = os.WriteFile(os.ExpandEnv(`$USERPROFILE\.dobby\gumroad_debug.txt`),
		[]byte(fmt.Sprintf("status=%d\n%s\n", resp.StatusCode, bodyBytes)), 0o600)

	var result gumroadVerifyResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return fmt.Errorf("無法連線至驗證伺服器，請確認網路連線: %w", err)
	}

	if !result.Success {
		return license.ErrInvalidLicenseKey
	}
	if result.Uses > 1 {
		return license.ErrLicenseAlreadyUsed
	}
	return nil
}
