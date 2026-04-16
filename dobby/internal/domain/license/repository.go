package license

import "context"

// ILicenseRepository is the persistence port for the License aggregate.
type ILicenseRepository interface {
	// Load reads the stored license. Returns (nil, false, nil) when no record exists yet.
	Load(ctx context.Context) (*License, bool, error)
	// Save persists the license, overwriting any existing record.
	Save(ctx context.Context, l *License) error
}
