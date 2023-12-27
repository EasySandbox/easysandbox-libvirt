package shootbacklauncher

import (
	_ "embed"
	"os"
	"path/filepath"
	"time"

	"git.voidnet.tech/kev/easysandbox/sandbox"
	"github.com/estebangarcia21/subprocess"
)

//go:embed shootback/master.py
var ShootbackMasterData []byte

//go:embed shootback/common_func.py
var ShootbackLibraryData []byte

//go:embed shootback/slaver.py
var ShootbackSlaverPyData []byte

func StartShootbackMaster(sandboxName string) (string, string, error) {
	var err error
	sandboxDir := filepath.Join(sandbox.SandboxInstallDir, sandboxName)

	masterFilePath := filepath.Join(sandboxDir, "master.py")
	err = os.WriteFile(masterFilePath, ShootbackMasterData, 0700)
	if err != nil {
		return "", "", err
	}

	libraryFilePath := filepath.Join(sandboxDir, "common_func.py")
	err = os.WriteFile(libraryFilePath, ShootbackLibraryData, 0644)
	if err != nil {
		return "", "", err
	}

	xpraPortFilePath := filepath.Join(sandboxDir, "xpra-port")
	vmPortFilePath := filepath.Join(sandboxDir, "vm-port")

	go func() {
		subprocess.New(
			masterFilePath,
			subprocess.Args(
				"-x", xpraPortFilePath,
				"-z", vmPortFilePath)).Exec()
	}()

	for {
		if _, err := os.Stat(vmPortFilePath); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	xpraPort, err := os.ReadFile(xpraPortFilePath)
	if err != nil {
		return "", "", err
	}

	vmPort, err := os.ReadFile(vmPortFilePath)
	if err != nil {
		return "", "", err
	}

	return string(xpraPort), string(vmPort), nil
}
