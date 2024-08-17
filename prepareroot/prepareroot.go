package prepareroot

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/estebangarcia21/subprocess"

	"git.voidnet.tech/kev/easysandbox/sandbox"
)

//go:embed templatedata/xpra.service
var xpraSystemdServiceData []byte
const xpraVsockPIDTag = "{VSOCK_ID}"


func PrepareRoot(sandboxName string) error {
	fmt.Println("Preparing root for sandbox", sandboxName)
	targetRootFile := sandbox.SandboxInstallDir + sandboxName + "/root.qcow2"

	xpraSystemdServiceFile, err := os.CreateTemp("", "xpra.service")

	if _, err := xpraSystemdServiceFile.WriteString(
		strings.Replace(
			string(xpraSystemdServiceData),
			xpraVsockPIDTag, fmt.Sprintf("%d", 4), 1)); err != nil {
		return fmt.Errorf("failed to write xpra systemd service file: %w", err)
	}
	defer func() {
		os.Remove(xpraSystemdServiceFile.Name())
	}()

	if err != nil {
		return fmt.Errorf("failed to setup xpra systemd service for VM: %w", err)
	}

	virtCustomizeArgs := subprocess.Args(
		"-a", targetRootFile,
		"--no-selinux-relabel",
		"--firstboot-command", "setenforce 0",
		"--root-password", "password:"+"pass",
		"--upload",
		fmt.Sprintf("%s:%s", xpraSystemdServiceFile.Name(), "/etc/systemd/system/xpra.service"),
		//"--append-line", "/etc/fstab:/dev/sdb1 /user/ ext4 defaults 0 1",
		"--firstboot-command", "systemctl daemon-reload",
		"--firstboot-command", "systemctl enable xpra.service",
		"--firstboot-command", "systemctl start xpra.service",
		"--delete", "/etc/ssh/*_key",
		"--delete", "/etc/ssh/*.pub",
		"--hostname", sandboxName,
	)

	customizeProc := subprocess.New("virt-customize", virtCustomizeArgs)
	err = customizeProc.Exec()

	if err != nil {
		return fmt.Errorf("failed to configure domain root: %w", err)
	}
	if customizeProc.ExitCode() != 0 {
		return fmt.Errorf("virt-customize exited with non-zero exit code: %d", customizeProc.ExitCode())
	}

	os.WriteFile(sandbox.SandboxInstallDir+sandboxName+"/root-prepared", []byte{}, 0644)

	return nil
}
