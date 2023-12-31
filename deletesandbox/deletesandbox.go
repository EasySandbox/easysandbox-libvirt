package deletesandbox

import (
	"fmt"
	"os"

	"git.voidnet.tech/kev/easysandbox/sandbox"
	"libvirt.org/go/libvirt"
)

func DeleteLibvirtDomain(domainName string) error {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	sboxDomain, lookupError := conn.LookupDomainByName(domainName)

	if lookupError != nil {
		lvErr := lookupError.(libvirt.Error)
		if lvErr.Code != libvirt.ERR_NO_DOMAIN {
			return fmt.Errorf("error looking up domain: %w", lookupError)
		}
	} else {
		destroyError := sboxDomain.Destroy()
		if destroyError != nil {
			return fmt.Errorf("error destroying libvirt sandbox domain: %w", destroyError)
		}
		domainUndefineError := sboxDomain.Undefine()
		if domainUndefineError != nil {
			return fmt.Errorf("error undefining libvirt sandbox domain: %w", domainUndefineError)

		}
	}
	return nil
}

func DeleteSandbox(sandboxName string) error {

	if deleteLibvirtErr := DeleteLibvirtDomain(sandboxName); deleteLibvirtErr != nil {
		return fmt.Errorf("error deleting libvirt sandbox: %w", deleteLibvirtErr)
	}
	return os.RemoveAll(sandbox.SandboxInstallDir + sandboxName)
}
