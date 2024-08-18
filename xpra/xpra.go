package xpra

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"git.voidnet.tech/kev/easysandbox-livbirt/prepareroot"
	"git.voidnet.tech/kev/easysandbox-livbirt/sandboxrunning"
	"git.voidnet.tech/kev/easysandbox/sandbox"
	"github.com/estebangarcia21/subprocess"
	"libvirt.org/go/libvirt"
)

var XPRA_BIN_NAME = "xpra"


func GetSandboxVSockID(sandboxName string) (string, error) {

	vsockID, readVsockIDErr := os.ReadFile(fmt.Sprintf(
		"%s/%s/%s",
		sandbox.SandboxInstallDir,
		sandboxName,
		prepareroot.VSockIDFileName))

	if readVsockIDErr != nil {
		return "", readVsockIDErr
	}

	return string(vsockID), nil
}

func getSandboxXPRAConnectionString(sandboxName string) (string, error) {

	sandboxVsockID, err := GetSandboxVSockID(sandboxName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("vsock://%s:8888", sandboxVsockID), nil

}

func isXpraAttached(sandboxName string) (bool, error) {
	connString, getConnStringErr := getSandboxXPRAConnectionString(sandboxName)
	if getConnStringErr != nil {
		return false, getConnStringErr
	}
	attachSemaphorePath := filepath.Join(sandbox.SandboxInstallDir, sandboxName, "xpra-attach")
	semaphoreContents, semaphoreReadErr := os.ReadFile(attachSemaphorePath)
	if semaphoreReadErr != nil {
		if os.IsNotExist(semaphoreReadErr) {
			return false, nil
		}
		return false, semaphoreReadErr
	}
	if string(semaphoreContents) == connString {
		return true, nil
	}
	return false, nil
}

func StartXpraClient(sandboxName string) error {

	conn, err := libvirt.NewConnect("qemu:///session")
	if err != nil {
		return fmt.Errorf("error connecting to libvirt: %w", err)
	}

	for {
		sandboxRunning, sandboxRunningErr := sandboxrunning.IsSandboxRunning(sandboxName, conn)
		if sandboxRunningErr != nil {
			return sandboxRunningErr
		}
		if sandboxRunning {
			break
		}
		time.Sleep(time.Millisecond * 10)
	}

	xpraAttached, xpraAttachedErr := isXpraAttached(sandboxName)
	if xpraAttachedErr != nil {
		return xpraAttachedErr
	}
	if xpraAttached {
		return errors.New("xpra is already attached")
	}

	connString, getConnStringErr := getSandboxXPRAConnectionString(sandboxName)
	if getConnStringErr != nil {
		return getConnStringErr
	}

	var waitForServerProc *subprocess.Subprocess
	serverUp := false
	for !serverUp {
		waitForServerProc = subprocess.New("xpra", subprocess.Args("control", connString, "hello"))
		waitForServerProc.Exec()
		fmt.Println("exit code", waitForServerProc.ExitCode())
		if waitForServerProc.ExitCode() == 0 {
			serverUp = true
		}
		time.Sleep(time.Millisecond * 20)

	}

	attachSemaphorePath := filepath.Join(sandbox.SandboxInstallDir, sandboxName, "xpra-attach")
	defer os.Remove(attachSemaphorePath)
	xpraAttachSemaphoreErr := os.WriteFile(attachSemaphorePath, []byte(connString), 0644)
	if xpraAttachSemaphoreErr != nil {
		return xpraAttachSemaphoreErr
	}

	return subprocess.New("xpra", subprocess.Args("attach", connString, "--splash=no", "--dpi=100")).Exec()
}

func RunXpraCommand(sandboxName string, args ...string) error {

	xpraAttached, xpraAttachedErr := isXpraAttached(sandboxName)
	if xpraAttachedErr != nil {
		return xpraAttachedErr
	}
	if !xpraAttached {
		return errors.New("xpra is not attached")
	}

	connString, getConnStringErr := getSandboxXPRAConnectionString(sandboxName)
	if getConnStringErr != nil {
		return getConnStringErr
	}

	return subprocess.New("xpra", subprocess.Args("control", connString, "start", strings.Join(args, " "))).Exec()

}
