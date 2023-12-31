package shootbacklauncher

import (
	"os"
	"path/filepath"

	"git.voidnet.tech/kev/easysandbox/sandbox"

	"strconv"
	"syscall"
)

func KillShootbackMaster(sandboxName string) error {
	shootbackMasterPid, shootbackMasterPidErr := os.ReadFile(filepath.Join(sandbox.SandboxInstallDir, sandboxName, "shootbackmaster.pid"))
	if shootbackMasterPidErr != nil {
		return shootbackMasterPidErr
	}

	pid, err := strconv.Atoi(string(shootbackMasterPid))
	if err != nil {
		return err
	}

	process, processErr := os.FindProcess(pid)
	if processErr != nil {
		return processErr
	}

	killErr := process.Signal(syscall.SIGTERM)
	if killErr != nil {
		return killErr
	}

	os.Remove(filepath.Join(sandbox.SandboxInstallDir, sandboxName, "shootbackmaster.pid"))
	os.Remove(filepath.Join(sandbox.SandboxInstallDir, sandboxName, "xpra-port"))
	os.Remove(filepath.Join(sandbox.SandboxInstallDir, sandboxName, "vm-port"))


	return nil
}
