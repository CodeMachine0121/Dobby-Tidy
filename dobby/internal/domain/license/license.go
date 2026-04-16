package license

import "time"

// License is the Aggregate Root that tracks trial state and activation.
type License struct {
	trialStartedAt time.Time
	activatedKey   string
	machineID      string
}

// NewLicense creates a fresh license record anchored at the given trial start time.
func NewLicense(trialStartedAt time.Time) *License {
	return &License{trialStartedAt: trialStartedAt}
}

// Reconstitute rebuilds a License from persisted data (used by repositories only).
func Reconstitute(trialStartedAt time.Time, activatedKey, machineID string) *License {
	return &License{
		trialStartedAt: trialStartedAt,
		activatedKey:   activatedKey,
		machineID:      machineID,
	}
}

// Accessors
func (l *License) TrialStartedAt() time.Time { return l.trialStartedAt }
func (l *License) ActivatedKey() string      { return l.activatedKey }
func (l *License) MachineID() string         { return l.machineID }
func (l *License) IsActivated() bool         { return l.activatedKey != "" }

// IsTrialActive returns true when the trial window has not yet elapsed and the license is not activated.
func (l *License) IsTrialActive(now time.Time) bool {
	if l.IsActivated() {
		return false
	}
	expiry := l.trialStartedAt.Add(TrialDays * 24 * time.Hour)
	return now.Before(expiry)
}

// CanRun returns true when the background processor is permitted to operate.
func (l *License) CanRun(now time.Time) bool {
	return l.IsActivated() || l.IsTrialActive(now)
}

// Status returns the current LicenseStatus.
func (l *License) Status(now time.Time) LicenseStatus {
	if l.IsActivated() {
		return LicenseStatusActivated
	}
	if l.IsTrialActive(now) {
		return LicenseStatusActive
	}
	return LicenseStatusExpired
}

// DaysRemaining returns the number of full days left in the trial.
// Returns -1 when already activated; 0 when expired.
func (l *License) DaysRemaining(now time.Time) int {
	if l.IsActivated() {
		return -1
	}
	expiry := l.trialStartedAt.Add(TrialDays * 24 * time.Hour)
	remaining := expiry.Sub(now)
	if remaining <= 0 {
		return 0
	}
	days := int(remaining.Hours() / 24)
	if days == 0 {
		return 1 // less than a full day but still active
	}
	return days
}

// Activate transitions the license to the activated state, binding it to the given machine.
func (l *License) Activate(key, machineID string) error {
	if l.IsActivated() {
		return ErrAlreadyActivated
	}
	l.activatedKey = key
	l.machineID = machineID
	return nil
}
