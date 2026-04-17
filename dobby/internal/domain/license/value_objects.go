package license

import "errors"

// LicenseStatus represents the current authorization state of the application.
type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "active"    // within 14-day free trial
	LicenseStatusExpired   LicenseStatus = "expired"   // trial over, not yet activated
	LicenseStatusActivated LicenseStatus = "activated" // paid and activated
)

// TrialDays is the number of days the free trial lasts.
const TrialDays = 14

var ErrAlreadyActivated = errors.New("license is already activated")
var ErrInvalidLicenseKey = errors.New("無效的 license key")
var ErrLicenseAlreadyUsed = errors.New("此 key 已在另一台機器上使用")
