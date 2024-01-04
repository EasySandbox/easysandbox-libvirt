package templates

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
)

var RootTemplateDir = xdg.DataHome + "/easysandbox/root-templates/"
var HomeTemplateDir = xdg.DataHome + "/easysandbox/home-templates/"

func getDiskFilesInDir(dir string) (names []string, err error) {
	file, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to open dir %s: %w", dir, err)
	}
	names, err = file.Readdirnames(0)
	var filteredNames []string
	if err != nil {
		return nil, fmt.Errorf("failed to read directory file names in %s: %w", dir, err)
	}
	for i := 0; i < len(names); i++ {
		if names[i][len(names[i])-6:] == ".qcow2" {
			filteredNames = append(filteredNames, names[i])
		}
	}

	return filteredNames, nil
}

func GetRootTemplatePaths() (paths []string, err error) {

	files, err := getDiskFilesInDir(RootTemplateDir)

	return files, err

}

func GetHomeTemplatePaths() (paths []string, err error) {

	files, err := getDiskFilesInDir(HomeTemplateDir)

	return files, err

}

func GetRootTemplatesList() ([]string, error) {
	var templateList []string
	templateList, err := GetRootTemplatePaths()
	if err != nil {
		return templateList, err
	}

	return templateList, nil

}

func GetHomeTemplatesList() ([]string, error) {
	var templateList []string
	templateList, err := GetHomeTemplatePaths()
	if err != nil {
		return templateList, err
	}

	return templateList, nil
}
