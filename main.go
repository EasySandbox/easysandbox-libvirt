package main

import (
	"git.voidnet.tech/kev/easysandbox-livbirt/createlinkedclone"
	"git.voidnet.tech/kev/easysandbox-livbirt/deletesandbox"
	"git.voidnet.tech/kev/easysandbox-livbirt/getavailablevsockid"
	"git.voidnet.tech/kev/easysandbox-livbirt/prepareroot"
	"git.voidnet.tech/kev/easysandbox-livbirt/sandboxrunning"
	"git.voidnet.tech/kev/easysandbox-livbirt/templates"
	"git.voidnet.tech/kev/easysandbox-livbirt/virtinstallargs"
	"git.voidnet.tech/kev/easysandbox-livbirt/xpra"
	"git.voidnet.tech/kev/easysandbox/sandbox"

	"github.com/estebangarcia21/subprocess"
	"libvirt.org/go/libvirt"

	"fmt"
	"os"
	"time"
)

func createDirectory(path string) (err error) {
	return os.MkdirAll(path, 0700)
}

func createLibvirtDomain(sandboxName string, libvirtConn *libvirt.Connect, userProvidedArgs ...string) error {
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: %s sandbox_name [virtinstall args]", os.Args[0])
	} else if len(os.Args) > 3 {
		userProvidedArgs = os.Args[3:]
	}

	_, err := getavailablevsockid.GetAvailableVSockID(libvirtConn)
	if err != nil {
		panic(fmt.Errorf("failed to get available vsock id: %w", err))
	}

	virtInstallArgs := virtinstallargs.GetVirtInstallArgs(sandboxName, userProvidedArgs...)

	// kind of a hack, because we could use the libvirt api, but that involves XML
	if virtInstallCmdErr := subprocess.New("virt-install", virtInstallArgs).Exec(); virtInstallCmdErr != nil {
		return fmt.Errorf("error running virt-install: %w", virtInstallCmdErr)
	}
	return nil
}

func StartSandbox(name string) error {
	var userProvidedArgs = os.Args[3:]
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	if prepareRootErr := prepareroot.PrepareRoot(name, conn); prepareRootErr != nil {
		return fmt.Errorf("error preparing root for sandbox: %w", prepareRootErr)
	}

	startLibvirtDomain := func() error {
		sboxDomain, err := conn.LookupDomainByName(name)
		if err != nil {
			return fmt.Errorf("error looking up sandbox in libvirt: %w", err)
		}

		if sandboxStartError := sboxDomain.Create(); sandboxStartError != nil {
			return fmt.Errorf("error starting libvirt domain: %w", sandboxStartError)
		}
		return nil
	}

	if sandboxIsRunning, err := sandboxrunning.IsSandboxRunning(name, conn); err != nil {
		return err
	} else {
		if sandboxIsRunning {
			return &sandbox.SandboxIsRunningError{
				Sandbox: name,
				Msg:     "cannot start sandbox that is already running",
			}
		}
	}

	if errDeleteLibvirtDomain := deletesandbox.DeleteLibvirtDomain(name); errDeleteLibvirtDomain != nil {
		return errDeleteLibvirtDomain
	}

	if libvirtSetupErr := createLibvirtDomain(name, conn, userProvidedArgs...); libvirtSetupErr != nil {
		return fmt.Errorf("error setting up libvirt domain: %w", libvirtSetupErr)
	}

	return startLibvirtDomain()
}

func StopSandbox(name string) error {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	var shutdownState libvirt.DomainState
	var shutdownStateErr error

	domain, err := conn.LookupDomainByName(name)
	if err != nil {
		return fmt.Errorf("error looking up sandbox in libvirt: %w", err)
	}

	var shutdownAttemptTime = time.Now().Unix()
	for shutdownState != libvirt.DOMAIN_SHUTOFF {
		shutdownState, _, shutdownStateErr = domain.GetState()

		if shutdownStateErr != nil {
			return fmt.Errorf("error getting sandbox domain state: %w", shutdownStateErr)
		}
		time.Sleep(50 * time.Millisecond)

		if time.Now().Unix()-shutdownAttemptTime > 5 {
			if shutdownErr := domain.Shutdown(); shutdownErr != nil {
				return fmt.Errorf("error shutting down libvirt domain %s: %w", name, shutdownErr)
			}
			shutdownAttemptTime = time.Now().Unix()
		}

	}

	if domainUndefineErr := domain.Undefine(); domainUndefineErr != nil {
		return fmt.Errorf("error undefining libvirt domain %s: %w", name, domainUndefineErr)
	}

	return nil
}

func CreateSandbox(sbox sandbox.SandboxInfo) error {
	directoryCreateError := createDirectory(sandbox.SandboxInstallDir + sbox.Name)
	if directoryCreateError != nil {
		return fmt.Errorf("error creating sandbox directory: %w", directoryCreateError)
	}

	linkedCloneCreationError := createlinkedclone.CreateLinkedCloneOfTemplate(sbox.RootTemplate, sbox.Name, true)
	if linkedCloneCreationError != nil {
		return fmt.Errorf("error creating linked clone of root template: %w", linkedCloneCreationError)
	}

	linkedCloneCreationError = createlinkedclone.CreateLinkedCloneOfTemplate(sbox.HomeTemplate, sbox.Name, false)
	if linkedCloneCreationError != nil {
		return fmt.Errorf("error creating linked clone of home template: %w", linkedCloneCreationError)
	}
	return nil

}

func DeleteSandbox(sandboxName string) error {
	StopSandbox(sandboxName)
	return deletesandbox.DeleteSandbox(sandboxName)
}

func GetRootTemplatesList() ([]string, error) {
	return templates.GetRootTemplatesList()
}

func GetHomeTemplatesList() ([]string, error) {
	return templates.GetHomeTemplatesList()
}

func GUIExecute(sandboxName string, command ...string) error {
	return xpra.RunXpraCommand(sandboxName, command...)
}

func GUIAttach(sandboxName string) error {
	return xpra.StartXpraClient(sandboxName)
}
