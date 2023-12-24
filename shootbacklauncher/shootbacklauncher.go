package shootbacklauncher

import (
	_ "embed"
	"os"
	"path/filepath"
	"time"

	"github.com/estebangarcia21/subprocess"
)

//go:embed shootback/master.py
var shootbackMasterData []byte

//go:embed shootback/common_func.py
var shootbackLibraryData []byte

func StartShootbackMaster(sandboxName string) error {
	tempDir, err := os.MkdirTemp("", "shootback")
	if err != nil {
		return err
	}

	masterFilePath := filepath.Join(tempDir, "master.py")
	err = os.WriteFile(masterFilePath, shootbackMasterData, 0700)
	if err != nil {
		return err
	}

	libraryFilePath := filepath.Join(tempDir, "common_func.py")
	err = os.WriteFile(libraryFilePath, shootbackLibraryData, 0644)
	if err != nil {
		return err
	}

	go func() {
		subprocess.New(
			masterFilePath,
			subprocess.Args(
				"-x", filepath.Join(tempDir, "xpra-port"),
				"-z", filepath.Join(tempDir, "vm-port"))).Exec()
	}()
	for {
		if _, err := os.Stat(filepath.Join(tempDir, "vm-port")); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	return nil

}
