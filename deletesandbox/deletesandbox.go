package deletesandbox

import (
	"fmt"
	"os"

	"git.voidnet.tech/kev/easysandbox/sandbox"
	"libvirt.org/go/libvirt"
)

func DeleteSandbox(sandboxName string) error {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	sboxDomain, lookupError := conn.LookupDomainByName(sandboxName)

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
		if domainUndefineError != nil{
			return fmt.Errorf("error undefining libvirt sandbox domain: %w", domainUndefineError)

		}
	}

	return os.RemoveAll(sandbox.SandboxInstallDir + sandboxName)
}
