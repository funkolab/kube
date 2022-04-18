package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/funkolab/kube/pkg/version"
	"github.com/manifoldco/promptui"
)

func Execute() {

	var result string
	var config Kubeconfig
	var filePath string

	var lFlag = flag.Bool("l", false, "Launch a new shell and set KUBECONFIG")
	var vFlag = flag.Bool("v", false, "Print the version of the plugin")
	flag.Parse()

	if *vFlag {
		fmt.Printf("build date:%s by: %s\nplatform: %s/%s\n", version.Date, version.BuiltBy, version.OsName, version.PlatformName)
		os.Exit(0)
	}

	dirname, err := os.UserHomeDir()
	check(err)

	kubieFolder := filepath.Join(dirname, ".kube/kubie")
	configFile := filepath.Join(dirname, ".kube/config")

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (info.Mode() & os.ModeCharDevice) == 0 {
		// data is being piped to stdin
		bytes, _ := ioutil.ReadAll(os.Stdin)

		if len(bytes) > 0 {
			// there is something on STDIN

			config.init(bytes)

			result = strings.TrimSuffix(config.CurrentContext, "-context")

			filePath = filepath.Join(dirname, ".kube/kubie", result+".yaml")
			f, err := os.Create(filePath)
			check(err)

			defer f.Close()

			w := bufio.NewWriter(f)

			_, err = w.WriteString(string(bytes))
			check(err)

			w.Flush()

		}
	} else {
		// stdin is from a terminal

		var fileList []string
		var fileListNames []string

		files, err := ioutil.ReadDir(kubieFolder)
		if err != nil {
			log.Fatal(err)
		}

		if len(files) == 0 {
			fmt.Println("No kubeconfig files found")
			return
		}

		for _, file := range files {
			if file.Mode().IsRegular() {
				ext := filepath.Ext(file.Name())
				if ext == ".yaml" || ext == ".yml" {
					fileList = append(fileList, file.Name())
					fileListNames = append(fileListNames, strings.TrimSuffix(file.Name(), ext))
				}
			}
		}

		prompt := promptui.Select{
			Label: "Select Cluster",
			Items: fileListNames,
		}

		index, _, err := prompt.Run()

		result = fileList[index]

		filePath = filepath.Join(kubieFolder, result)

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		result = fileListNames[index]
	}

	if *lFlag {
		os.Setenv("KUBECONFIG", filePath)
		fmt.Printf("You set %q only for this session\n", result)
		syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, syscall.Environ())
	} else {
		err = copyFile(filePath, configFile)
		check(err)
		fmt.Printf("You set %q globally\n", result)
	}
}
