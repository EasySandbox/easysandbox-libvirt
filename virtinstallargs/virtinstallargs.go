package virtinstallargs

import (
	"fmt"
	"runtime"

	"github.com/estebangarcia21/subprocess"

	"git.voidnet.tech/kev/easysandbox-livbirt/xpra"
	"git.voidnet.tech/kev/easysandbox/sandbox"
)

func GetVirtInstallArgs(sandboxName string, args ...string) subprocess.Option {
	return subprocess.Args(GetVirtInstallArgsString(sandboxName, args...)...)
}

func GetVirtInstallArgsString(sandboxName string, args ...string) []string {
	rootCloneFile := sandbox.SandboxInstallDir + sandboxName + "/" + "root.qcow2"
	homeFile := sandbox.SandboxInstallDir + sandboxName + "/" + "home.qcow2"

	vsockID, err := xpra.GetSandboxVSockID(sandboxName)
	if err != nil {
		panic(fmt.Errorf("failed to get vsock id for sandbox:%w))", err))
	}

	mandatoryArgs := []string{
		"--name", sandboxName,
		"--disk", rootCloneFile + ",target.bus=sata",
		"--disk", homeFile + ",target.bus=sata",
		"--import",
		"--hvm",
		"--vsock", "cid=" + vsockID,
		"--virt-type", "kvm",
		"--install", "no_install=yes",
		"--noreboot",
	}
	defaultArgs := map[string]string{
		"--network":    "user",
		"--memory":     "4096",
		"--vcpus":      fmt.Sprintf("%d", runtime.NumCPU()),
		"--os-variant": "linux2022",
	}

	var overriddenArgs []string
	for i := 0; i < len(args); i += 2 {
		arg := args[i]
		value := args[i+1]
		delete(defaultArgs, arg)
		overriddenArgs = append(overriddenArgs, arg, value)
	}

	allArgs := append(mandatoryArgs, overriddenArgs...)
	for arg, value := range defaultArgs {
		allArgs = append(allArgs, arg, value)
	}

	return allArgs
}
