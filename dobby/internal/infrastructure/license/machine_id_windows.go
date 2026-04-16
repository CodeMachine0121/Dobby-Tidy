//go:build windows

package infralicense

import (
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/sys/windows/registry"
)

// WindowsMachineIdProvider reads the Windows cryptographic machine GUID from the registry
// and derives a stable 32-char hex identifier.
type WindowsMachineIdProvider struct{}

func NewMachineIdProvider() *WindowsMachineIdProvider {
	return &WindowsMachineIdProvider{}
}

func (p *WindowsMachineIdProvider) MachineID() (string, error) {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Cryptography`,
		registry.QUERY_VALUE|registry.WOW64_64KEY,
	)
	if err != nil {
		return fallbackMachineID(), nil
	}
	defer k.Close()

	guid, _, err := k.GetStringValue("MachineGuid")
	if err != nil {
		return fallbackMachineID(), nil
	}

	h := sha256.Sum256([]byte("dobby-machine-" + guid))
	return hex.EncodeToString(h[:16]), nil
}

func fallbackMachineID() string {
	return "windows-machine-fallback"
}
