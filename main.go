package main

import (
	"git.voidnet.tech/kev/easysandbox-livbirt/createlinkedclone"
	"git.voidnet.tech/kev/easysandbox-livbirt/deletesandbox"
	"git.voidnet.tech/kev/easysandbox-livbirt/prepareroot"
	"git.voidnet.tech/kev/easysandbox-livbirt/shootbacklauncher"
	"git.voidnet.tech/kev/easysandbox-livbirt/virtinstallargs"
	"git.voidnet.tech/kev/easysandbox/sandbox"

	"github.com/estebangarcia21/subprocess"
	"libvirt.org/go/libvirt"

	"errors"
	"fmt"
	"os"
	"time"
)

func createDirectory(path string) (err error) {
	return os.MkdirAll(path, 0700)
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

var rootPreparedSempaphoreFile = "root-prepared"


func prepVM(vmName string) (error) {
	_, vmPort, err := shootbacklauncher.StartShootbackMaster(vmName)
	if err != nil {
		return fmt.Errorf("could not prepare root for %s: %w", vmName, err)
	}
	return prepareroot.PrepareRoot(vmName, vmPort)
}

func StartSandbox(name string) error {
	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	targetSandbox, lookupError := conn.LookupDomainByName(name)
	if lookupError == nil {
		isActive, isActiveErr := targetSandbox.IsActive()
		if isActiveErr != nil {
			return fmt.Errorf("error checking if sandbox is active: %w", isActiveErr)
		}
		if isActive {
			return &sandbox.SandboxIsRunningError{
				Sandbox: name,
				Msg:     "cannot start sandbox that is already running",
			}
		}
	}

	var userProvidedArgs []string
	if len(os.Args) < 3 {
		return fmt.Errorf("usage: %s sandbox_name [virtinstall args]", os.Args[0])
	} else if len(os.Args) > 3 {
		userProvidedArgs = os.Args[3:]
	}
	semaphorePath := fmt.Sprintf("%s/%s", sandbox.SandboxInstallDir+name, rootPreparedSempaphoreFile)

	exists, semaphoreExistsErr := pathExists(semaphorePath)
	if semaphoreExistsErr != nil {
		return fmt.Errorf("failed to check for root status file for %s: %w", name, semaphoreExistsErr)
	}
	if !exists {
		prepVMErr := prepVM(name)
		if prepVMErr != nil {
			return fmt.Errorf("failed to prepare VM for %s: %w", name, prepVMErr)
		}
	}

	sboxDomain, lookupError := conn.LookupDomainByName(name)

	if lookupError != nil {
		lvErr := lookupError.(libvirt.Error)
		if lvErr.Code != libvirt.ERR_NO_DOMAIN {
			return fmt.Errorf("error looking up domain: %w", lookupError)
		}
		virtInstallArgs := virtinstallargs.GetVirtInstallArgs(name, userProvidedArgs...)

		// kind of a hack, because we could use the libvirt api, but that involves XML
		if virtInstallCmdErr := subprocess.New("virt-install", virtInstallArgs).Exec(); virtInstallCmdErr != nil {
			return fmt.Errorf("error running virt-install: %w", virtInstallCmdErr)
		}
		sboxDomain, lookupError = conn.LookupDomainByName(name)
		if lookupError != nil {
			return fmt.Errorf("error looking up domain after libvirt definition: %w", lookupError)
		}
	}

	if sandboxStartError := sboxDomain.Create(); sandboxStartError != nil {
		return fmt.Errorf("error starting libvirt domain: %w", sandboxStartError)
	}
	return nil
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
	// create linked clone of root
	// create linked clone of home template (parent changes do not propagate)
	// libvirt install a machine with

	//fmt.Println("Not implemented")

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
	return deletesandbox.DeleteSandbox(sandboxName)
}
