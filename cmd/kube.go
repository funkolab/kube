package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  kube [flags] [filter]\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if flag.NArg() > 1 {
		fmt.Println("too may arguments")
		flag.Usage()
		os.Exit(1)
	}

	if *vFlag {
		fmt.Printf("kube: version %s (%s)\n", version.Version, version.Commit)
		fmt.Printf("build date:%s by: %s\nplatform: %s/%s\n", version.Date, version.BuiltBy, version.OsName, version.PlatformName)
		os.Exit(0)
	}

	dirname, err := os.UserHomeDir()
	check(err)

	kubieFolder := filepath.Join(dirname, ".kube/kubie")
	configFile := filepath.Join(dirname, ".kube/config")

	if isInputFromPipe() {
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
						if flag.NArg() == 1 && !strings.Contains(name, flag.Arg(0)) {
							continue
						}
						selectList = append(selectList, kubeChoice{Name: name, Context: context, Config: config, filePath: filePath})
					}

				}
			}
		}

		if len(selectList) == 0 {
			fmt.Printf("No cluster found with the filter \"%s\"\n", flag.Arg(0))
			return
		}

		var index = 0
		if len(selectList) > 1 {

			index, err = fuzzyfinder.Find(
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
		}

		result = selectList[index].Name
		filePath = selectList[index].filePath
		config = selectList[index].Config
	}

	if *lFlag {
		os.Setenv("KUBECONFIG", filePath)
		fmt.Printf("You set %q only for this session\n", result)
		cmd := exec.Command(os.Getenv("SHELL"))
		cmd.Stdin, _ = os.Open("/dev/tty")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		check(err)
		fmt.Printf("You disconnect from %q\n", result)
	} else {

		if result != "" {
			config.CurrentContext = result
		}

		err = clientcmd.WriteToFile(*config, configFile)
		check(err)

		fmt.Printf("You set %q globally\n", result)
	}
}
