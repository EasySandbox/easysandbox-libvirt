package sandboxrunning

import (
	"fmt"

	"libvirt.org/go/libvirt"
)

func IsSandboxRunning(name string, libvirtConn *libvirt.Connect) (bool, error) {
	targetSandbox, lookupError := libvirtConn.LookupDomainByName(name)
	if lookupError == nil {
		isActive, isActiveErr := targetSandbox.IsActive()
		if isActiveErr != nil {
			return false, fmt.Errorf("error checking if sandbox is active: %w", isActiveErr)
		}
		if isActive {
			return true, nil
		}
		return false, nil
	} else {
		lvErr := lookupError.(libvirt.Error)
		if lvErr.Code != libvirt.ERR_NO_DOMAIN {
			return false, fmt.Errorf("error looking up domain: %w", lookupError)
		}
		return false, nil
	}
}
