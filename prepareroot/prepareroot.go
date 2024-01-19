package prepareroot

import (
	"bytes"
	_ "embed"
	"fmt"
	"net"
	"os"

	"github.com/estebangarcia21/subprocess"

	"git.voidnet.tech/kev/easysandbox-livbirt/shootbacklauncher"
	"git.voidnet.tech/kev/easysandbox/sandbox"
)

//go:embed templatedata/shootback.service
var shootbackShootbackServiceData []byte

//go:embed templatedata/xpra.service
var xpraSystemdServiceData []byte

func getOurIP() (string, error) {
	conn, err := net.Dial("udp", "1.1.1.1:80") // placeholder IP address, no network activity is actually done
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func PrepareRoot(sandboxName string, vmShootbackPort string) error {
	targetRootFile := sandbox.SandboxInstallDir + sandboxName + "/root.qcow2"

	var err error

	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	hostIP, err := getOurIP()
	if err != nil {
		return fmt.Errorf("faiiled to get host IP: %w", err)
	}

	hostShootbackAddress := fmt.Sprintf("%s:%s", hostIP, vmShootbackPort)

	shootbackShootbackServiceFile, err := os.CreateTemp("", "shootbace.service")
	if err != nil {
		return fmt.Errorf("failed to create shootback systemd service file: %w", err)
	}
	defer func() {
		os.Remove(shootbackShootbackServiceFile.Name())
	}()

	fmt.Println("setting host_shootback to" + hostShootbackAddress)

	shootbackServiceData := bytes.ReplaceAll(shootbackShootbackServiceData, []byte("{host_shootback_address}"), []byte(hostShootbackAddress))

	if _, err := shootbackShootbackServiceFile.Write(shootbackServiceData); err != nil {
		return fmt.Errorf("failed to write shootback systemd service file: %w", err)
	}

	shootbackSlaverPyFile, err := os.CreateTemp("", "slaver.py")
	if err != nil {
		return fmt.Errorf("failed to create shootback slaver.py file: %w", err)
	}
	defer func() {
		os.Remove(shootbackSlaverPyFile.Name())
	}()

	if _, err := shootbackSlaverPyFile.Write(shootbacklauncher.ShootbackSlaverPyData); err != nil {
		return fmt.Errorf("failed to write shootback slaver.py file: %w", err)
	}

	shootbackLibraryFile, err := os.CreateTemp("", "common_func.py")
	if err != nil {
		return fmt.Errorf("failed to create shootback common_func.py file: %w", err)
	}
	defer func() {
		os.Remove(shootbackLibraryFile.Name())
	}()

	if _, err := shootbackLibraryFile.Write(shootbacklauncher.ShootbackLibraryData); err != nil {
		return fmt.Errorf("failed to write shootback common_func.py file: %w", err)
	}

	xpraSystemdServiceFile, err := os.CreateTemp("", "xpra.service")

	if _, err := xpraSystemdServiceFile.Write(xpraSystemdServiceData); err != nil {
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
		"--mkdir", "/shootback/",
		"--upload",
		fmt.Sprintf("%s:%s", xpraSystemdServiceFile.Name(), "/etc/systemd/system/xpra.service"),
		"--upload",
		fmt.Sprintf("%s:%s", shootbackShootbackServiceFile.Name(), "/etc/systemd/system/shootback.service"),
		"--upload",
		fmt.Sprintf("%s:%s", shootbackSlaverPyFile.Name(), "/shootback/slaver.py"),
		"--upload",
		fmt.Sprintf("%s:%s", shootbackLibraryFile.Name(), "/shootback/common_func.py"),
		"--chmod", "700:/shootback/slaver.py",
		//"--append-line", "/etc/fstab:/dev/sdb1 /user/ ext4 defaults 0 1",
		"--append-line", fmt.Sprintf("/etc/hosts:%s hostsystem", hostIP),
		"--firstboot-command", "systemctl daemon-reload",
		"--firstboot-command", "systemctl enable xpra.service",
		"--firstboot-command", "systemctl start xpra.service",
		"--firstboot-command", "systemctl enable shootback.service",
		"--firstboot-command", "systemctl start shootback.service",
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
