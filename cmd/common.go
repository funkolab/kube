package cmd

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isInputFromPipe() bool {
	fileInfo, err := os.Stdin.Stat()
	check(err)
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func launchShell(choice *kubeChoice) {
	os.Setenv("KUBECONFIG", choice.filePath)
	fmt.Printf("You set %q only for this session\n", choice.Name)
	cmd := exec.Command(os.Getenv("SHELL"))
	cmd.Stdin, _ = os.Open("/dev/tty")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	check(err)
	fmt.Printf("You disconnect from %q\n", choice.Name)
}

func buildList(files []fs.FileInfo, kubieFolder string) []kubeChoice {

	var selectList []kubeChoice

	for _, file := range files {
		if file.Mode().IsRegular() {
			ext := filepath.Ext(file.Name())
			if ext == ".yaml" || ext == ".yml" {

				filePath := filepath.Join(kubieFolder, file.Name())
				config := clientcmd.GetConfigFromFileOrDie(filePath)

				for name, context := range config.Contexts {
					if flag.NArg() == 1 && !strings.Contains(name, flag.Arg(0)) {
						continue
					}
					selectList = append(selectList, kubeChoice{Name: name, ContextName: name, Context: context, Config: config, filePath: filePath})
				}

			}
		}
	}
	return selectList
}
