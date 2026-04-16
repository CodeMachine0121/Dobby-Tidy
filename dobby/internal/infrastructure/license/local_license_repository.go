package infralicense

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/dobby/filemanager/internal/domain/license"
)

// obfuscationKey is used for XOR obfuscation of the license file.
// This deters casual inspection or simple file-deletion exploits; not cryptographic security.
const obfuscationKey = "dobby-license-store-xor-key"

type licenseRecord struct {
	TrialStartedAt string `json:"t"`
	ActivatedKey   string `json:"k,omitempty"`
	MachineID      string `json:"m,omitempty"`
}

// LocalLicenseRepository persists the License aggregate to an obfuscated local file.
type LocalLicenseRepository struct {
	filePath string
}

func NewLocalLicenseRepository(dataDir string) *LocalLicenseRepository {
	return &LocalLicenseRepository{
		filePath: filepath.Join(dataDir, "license.dat"),
	}
}

// Load reads and decodes the license file. Returns (nil, false, nil) if no file exists yet.
func (r *LocalLicenseRepository) Load(_ context.Context) (*license.License, bool, error) {
	raw, err := os.ReadFile(r.filePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	plain, err := deobfuscate(raw)
	if err != nil {
		return nil, false, err
	}
	var rec licenseRecord
	if err := json.Unmarshal(plain, &rec); err != nil {
		return nil, false, err
	}
	t, err := time.Parse(time.RFC3339, rec.TrialStartedAt)
	if err != nil {
		return nil, false, err
	}
	return license.Reconstitute(t, rec.ActivatedKey, rec.MachineID), true, nil
}

// Save encodes and writes the license to disk.
func (r *LocalLicenseRepository) Save(_ context.Context, l *license.License) error {
	rec := licenseRecord{
		TrialStartedAt: l.TrialStartedAt().UTC().Format(time.RFC3339),
		ActivatedKey:   l.ActivatedKey(),
		MachineID:      l.MachineID(),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.filePath), 0o700); err != nil {
		return err
	}
	return os.WriteFile(r.filePath, obfuscate(data), 0o600)
}

func obfuscate(data []byte) []byte {
	key := []byte(obfuscationKey)
	xored := make([]byte, len(data))
	for i, b := range data {
		xored[i] = b ^ key[i%len(key)]
	}
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(xored)))
	base64.StdEncoding.Encode(encoded, xored)
	return encoded
}

func deobfuscate(data []byte) ([]byte, error) {
	xored := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(xored, data)
	if err != nil {
		return nil, err
	}
	xored = xored[:n]
	key := []byte(obfuscationKey)
	plain := make([]byte, len(xored))
	for i, b := range xored {
		plain[i] = b ^ key[i%len(key)]
	}
	return plain, nil
}
