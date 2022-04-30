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
	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type kubeChoice struct {
	Name     string
	Context  *clientcmdapi.Context
	Config   *clientcmdapi.Config
	filePath string
}

func Execute() {

	var result string
	var config *clientcmdapi.Config
	var filePath string

	var lFlag = flag.Bool("l", false, "Launch a new shell and set KUBECONFIG")
	var vFlag = flag.Bool("v", false, "Print the version of the plugin")
	flag.Parse()

	if *vFlag {
		fmt.Printf("kube: version %s (%s)\n", version.Version, version.Commit)
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

			config, err = clientcmd.Load(bytes)
			if err != nil {
				fmt.Printf("Kubeconfig format is invalid: %s\n", err)
				os.Exit(1)
			}

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
		// interactive mode

		var selectList []kubeChoice

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

					filePath = filepath.Join(kubieFolder, file.Name())
					config = clientcmd.GetConfigFromFileOrDie(filePath)

					for name, context := range config.Contexts {
						selectList = append(selectList, kubeChoice{Name: name, Context: context, Config: config, filePath: filePath})
					}

				}
			}
		}

		index, err := fuzzyfinder.Find(
			selectList,
			func(i int) string {
				return selectList[i].Name
			},
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i == -1 {
					return ""
				}

				context := strings.ToUpper(strings.TrimSuffix(selectList[i].Name, "-context"))
				server := strings.TrimPrefix(selectList[i].Config.Clusters[selectList[i].Context.Cluster].Server, "https://")
				user := selectList[i].Context.AuthInfo
				namespace := selectList[i].Context.Namespace
				if namespace == "" {
					namespace = "default"
				}

				return fmt.Sprintf("%s\n\n    server:  %s\n      user:  %s\n namespace:  %s", context, server, user, namespace)
			}),
		)
		if err != nil {
			fmt.Printf("Selection %s\n", err)
			os.Exit(1)
		}

		result = selectList[index].Name
		filePath = selectList[index].filePath
		config = selectList[index].Config
	}

	if *lFlag {
		os.Setenv("KUBECONFIG", filePath)
		fmt.Printf("You set %q only for this session\n", result)
		syscall.Exec(os.Getenv("SHELL"), []string{os.Getenv("SHELL")}, syscall.Environ())
	} else {

		if result != "" {
			config.CurrentContext = result
		}

		err = clientcmd.WriteToFile(*config, configFile)
		check(err)

		fmt.Printf("You set %q globally\n", result)
	}
}
