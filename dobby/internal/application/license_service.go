package application

import (
	"context"
	"strings"
	"time"

	"github.com/dobby/filemanager/internal/domain/license"
)

// IGumroadVerifier verifies a Gumroad license key against the Gumroad API.
type IGumroadVerifier interface {
	Verify(ctx context.Context, licenseKey string) error
}

// IMachineIdProvider abstracts the platform-specific machine identifier.
type IMachineIdProvider interface {
	MachineID() (string, error)
}

// LicenseInfo is the read model returned to the presentation layer.
type LicenseInfo struct {
	Status        string `json:"status"`        // "active" | "expired" | "activated"
	DaysRemaining int    `json:"daysRemaining"` // -1 when activated, 0 when expired
}

// LicenseService orchestrates trial initialization, status queries, and license activation.
type LicenseService struct {
	repo      license.ILicenseRepository
	verifier  IGumroadVerifier
	machineId IMachineIdProvider
}

func NewLicenseService(repo license.ILicenseRepository, machineId IMachineIdProvider, verifier IGumroadVerifier) *LicenseService {
	return &LicenseService{
		repo:      repo,
		verifier:  verifier,
		machineId: machineId,
	}
}

// InitializeTrial records the trial start date on first launch. Idempotent if already initialized.
func (s *LicenseService) InitializeTrial(ctx context.Context) error {
	_, found, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}
	if found {
		return nil
	}
	return s.repo.Save(ctx, license.NewLicense(time.Now()))
}

// GetLicenseInfo returns the current license status and remaining trial days.
func (s *LicenseService) GetLicenseInfo(ctx context.Context) (*LicenseInfo, error) {
	l, found, err := s.repo.Load(ctx)
	if err != nil {
		return nil, err
	}
	if !found {
		return &LicenseInfo{Status: string(license.LicenseStatusExpired), DaysRemaining: 0}, nil
	}
	now := time.Now()
	return &LicenseInfo{
		Status:        string(l.Status(now)),
		DaysRemaining: l.DaysRemaining(now),
	}, nil
}

// ActivateLicense verifies the key via Gumroad API and, if valid, activates the license for this machine.
func (s *LicenseService) ActivateLicense(ctx context.Context, key string) error {
	key = strings.TrimSpace(key)
	l, _, err := s.repo.Load(ctx)
	if err != nil {
		return err
	}
	if l != nil && l.IsActivated() {
		return license.ErrAlreadyActivated
	}
	if err := s.verifier.Verify(ctx, key); err != nil {
		return err
	}
	machineID, err := s.machineId.MachineID()
	if err != nil {
		return err
	}
	if l == nil {
		l = license.NewLicense(time.Now())
	}
	if err := l.Activate(key, machineID); err != nil {
		return err
	}
	return s.repo.Save(ctx, l)
}

// CanRunBackgroundProcessor returns true when the current license permits the processor to run.
func (s *LicenseService) CanRunBackgroundProcessor(ctx context.Context) (bool, error) {
	l, found, err := s.repo.Load(ctx)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return l.CanRun(time.Now()), nil
}
