package createlinkedclone

import (
	"git.voidnet.tech/kev/easysandbox-livbirt/templates"
	"git.voidnet.tech/kev/easysandbox/sandbox"

	"github.com/estebangarcia21/subprocess"

	"strings"
)

func CreateLinkedCloneOfTemplate(template string, name string, isRoot bool) error {
	// ensure template ends with .qcow2
	if !strings.HasSuffix(template, ".qcow2") {
		template += ".qcow2"
	}

	var templatePath string
	var targetFile string
	if isRoot {
		templatePath = templates.RootTemplateDir + template
		targetFile = sandbox.SandboxInstallDir + name + "/root.qcow2"
	} else {
		templatePath = templates.HomeTemplateDir + template

		targetFile = sandbox.SandboxInstallDir + name + "/home.qcow2"
	}

	return subprocess.New(
		"qemu-img",
		subprocess.Arg("create"),
		subprocess.Arg("-f"),
		subprocess.Arg("qcow2"),
		subprocess.Arg("-F"),
		subprocess.Arg("qcow2"),
		subprocess.Arg("-b"),
		subprocess.Arg(templatePath),
		subprocess.Arg(targetFile)).Exec()
}
