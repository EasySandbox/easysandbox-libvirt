package prepareroot

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/estebangarcia21/subprocess"

	"git.voidnet.tech/kev/easysandbox/generatepassword"
	"git.voidnet.tech/kev/easysandbox/getourip"
	"git.voidnet.tech/kev/easysandbox/sandbox"
)


//go:embed templatedata/ipmapper.sh
var ipmapperClientScriptData []byte

//go:embed templatedata/ipmapper.service
var ipmapperSystemdServiceData []byte

//go:embed templatedata/ipmapper.timer
var ipmapperSystemdTimerData []byte

//go:embed templatedata/xpra.service
var xpraSystemdServiceData []byte

func PrepareRoot(sandboxName string) error {
	targetRootFile := sandbox.SandboxInstallDir + sandboxName + "/root.qcow2"

	var err error
	pass, err := generatepassword.GenerateRandomPassword(12)
	pass = "password"

	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	hostIP, err := getourip.GetOurIP()
	if err != nil {
		return fmt.Errorf("faiiled to get host IP: %w", err)
	}

	ipmapperSystemdServiceFile, err := os.CreateTemp("", "ipmapper.service")

	ipmapperSystemdServiceFile.Write(ipmapperSystemdServiceData)

	defer func() {
		os.Remove(ipmapperSystemdServiceFile.Name())
	}()

	if err != nil {
		return fmt.Errorf("failed to setup systemd service file for VM: %w", err)
	}

	ipmapperSystemdTimerFile, err := os.CreateTemp("", "ipmapper.timer")

	ipmapperSystemdTimerFile.Write(ipmapperSystemdTimerData)

	defer func() {
		os.Remove(ipmapperSystemdTimerFile.Name())
	}()

	if err != nil {
		return fmt.Errorf("failed to setup systemd timer file for VM: %w", err)
	}

	ipmapperClientScriptFile, err := os.CreateTemp("", "ipmapper.sh")

	ipmapperClientScriptFile.Write(ipmapperClientScriptData)

	defer func() {
		os.Remove(ipmapperClientScriptFile.Name())
	}()

	if err != nil {
		return fmt.Errorf("failed to setup ipmapper script for VM: %w", err)
	}

	xpraSystemdServiceFile, err := os.CreateTemp("", "xpra.service")

	xpraSystemdServiceFile.Write(xpraSystemdServiceData)

	defer func() {
		os.Remove(xpraSystemdServiceFile.Name())
	}()

	if err != nil {
		return fmt.Errorf("failed to setup xpra systemd service for VM: %w", err)
	}

	customizeProc := subprocess.New(
		"virt-customize",
		subprocess.Args(
			"-a", targetRootFile,
			"--no-selinux-relabel",
			"--firstboot-command", "setenforce 0",
			"--root-password", "password:"+pass,
			"--upload",
			fmt.Sprintf("%s:%s", ipmapperClientScriptFile.Name(), "/bin/ipmapper.sh"),
			"--chmod", "755:/bin/ipmapper.sh",
			"--upload",
			fmt.Sprintf("%s:%s", ipmapperSystemdServiceFile.Name(), "/etc/systemd/system/ipmapper.service"),
			"--upload",
			fmt.Sprintf("%s:%s", ipmapperSystemdTimerFile.Name(), "/etc/systemd/system/ipmapper.timer"),
			"--upload",
			fmt.Sprintf("%s:%s", xpraSystemdServiceFile.Name(), "/etc/systemd/system/xpra.service"),
			"--append-line", "/etc/fstab:/dev/sdb1 /user/ ext4 defaults 0 1",
			"--append-line", fmt.Sprintf("/etc/hosts:%s hostsystem", hostIP),
			"--firstboot-command", "systemctl daemon-reload",
			"--firstboot-command", "systemctl enable ipmapper.timer",
			"--firstboot-command", "systemctl start ipmapper.timer",
			"--firstboot-command", "systemctl enable xpra.service",
			"--firstboot-command", "systemctl start xpra.service",
			"--delete", "/etc/ssh/*_key",
			"--delete", "/etc/ssh/*.pub",
			"--hostname", sandboxName,
		),
	)
	err = customizeProc.Exec()

	if err != nil {
		fmt.Println("RAISING ERROR")
		return fmt.Errorf("failed to configure domain root: %w", err)
	}
	if customizeProc.ExitCode() != 0 {
		return fmt.Errorf("virt-customize exited with non-zero exit code: %d", customizeProc.ExitCode())
	}
	fmt.Println("Password for root: " + pass)

	os.WriteFile(sandbox.SandboxInstallDir+sandboxName+"/root-prepared", []byte{}, 0644)

	return nil
}
