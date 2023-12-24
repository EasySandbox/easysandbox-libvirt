package prepareroot

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/estebangarcia21/subprocess"

	"git.voidnet.tech/kev/easysandbox/getourip"
	"git.voidnet.tech/kev/easysandbox/sandbox"
)


//go:embed templatedata/xpra.service
var xpraSystemdServiceData []byte

func PrepareRoot(sandboxName string) error {
	targetRootFile := sandbox.SandboxInstallDir + sandboxName + "/root.qcow2"

	var err error

	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	hostIP, err := getourip.GetOurIP()
	if err != nil {
		return fmt.Errorf("faiiled to get host IP: %w", err)
	}

	xpraSystemdServiceFile, err := os.CreateTemp("", "xpra.service")

	xpraSystemdServiceFile.Write(xpraSystemdServiceData)

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
		//"--write", "/tmp/easysandboxguiport:",
		"--upload",
		fmt.Sprintf("%s:%s", xpraSystemdServiceFile.Name(), "/etc/systemd/system/xpra.service"),
		"--append-line", "/etc/fstab:/dev/sdb1 /user/ ext4 defaults 0 1",
		"--append-line", fmt.Sprintf("/etc/hosts:%s hostsystem", hostIP),
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
	//fmt.Println("Password for root: " + pass)

	os.WriteFile(sandbox.SandboxInstallDir+sandboxName+"/root-prepared", []byte{}, 0644)

	return nil
}
