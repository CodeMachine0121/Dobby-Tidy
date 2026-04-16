//go:build !windows

package infralicense

import "runtime"

// WindowsMachineIdProvider is a stub for non-Windows platforms (CI / dev builds).
type WindowsMachineIdProvider struct{}

func NewMachineIdProvider() *WindowsMachineIdProvider {
	return &WindowsMachineIdProvider{}
}

func (p *WindowsMachineIdProvider) MachineID() (string, error) {
	return "dev-" + runtime.GOOS, nil
}
