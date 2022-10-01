package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/funkolab/kube/pkg/version"
	"github.com/ktr0731/go-fuzzyfinder"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type kubeChoice struct {
	Name        string
	ContextName string
	Context     *clientcmdapi.Context
	Config      *clientcmdapi.Config
	filePath    string
}

func ProcessFromPipe() *kubeChoice {
	bytes, _ := io.ReadAll(os.Stdin)

	// test if there is something on STDIN
	if len(bytes) > 0 {

		config, err := clientcmd.Load(bytes)
		if err != nil {
			fmt.Printf("Kubeconfig format is invalid: %s\n", err)
			os.Exit(1)
		}

		name := strings.TrimSuffix(config.CurrentContext, "-context")

		dirname, err := os.UserHomeDir()
		check(err)

		filePath := filepath.Join(dirname, ".kube/kubie", name+".yaml")
		f, err := os.Create(filePath)
		check(err)

		defer f.Close()

		err = os.Chmod(filePath, 0600)
		check(err)

		w := bufio.NewWriter(f)

		_, err = w.WriteString(string(bytes))
		check(err)

		w.Flush()

		return &kubeChoice{
			Name:        name,
			ContextName: config.CurrentContext,
			Context:     config.Contexts[config.CurrentContext],
			Config:      config,
			filePath:    filePath,
		}

	}
	return nil
}

func InteractiveSelect() *kubeChoice {

	selectList := buildList()

	checkList(&selectList)

	var index = 0
	var err error

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

	return &selectList[index]
}

func InteractiveDelete() {
	listToClean := buildList()
	checkList(&listToClean)

	index, err := fuzzyfinder.FindMulti(
		listToClean,
		func(i int) string {
			return listToClean[i].Name
		},
	)

	if err != nil {
		fmt.Printf("Selection %s\n", err)
		os.Exit(1)
	}

	for _, choice := range index {
		os.Remove(listToClean[choice].filePath)
	}
}

func Execute() {

	var choice *kubeChoice

	var lFlag = flag.Bool("l", false, "Launch a new shell and set KUBECONFIG")
	var vFlag = flag.Bool("v", false, "Print the version of the plugin")
	var dFlag = flag.Bool("d", false, "Clean the kubeconfig files")

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

	if *dFlag {
		InteractiveDelete()
		os.Exit(0)
	}

	if isInputFromPipe() {
		choice = ProcessFromPipe()
	} else {
		choice = InteractiveSelect()
	}

	if *lFlag {
		launchShell(choice)
	} else {

		choice.Config.CurrentContext = choice.ContextName

		dirname, err := os.UserHomeDir()
		check(err)
		configFile := filepath.Join(dirname, ".kube/config")
		err = clientcmd.WriteToFile(*choice.Config, configFile)
		check(err)

		fmt.Printf("You set %q globally\n", choice.Name)
	}
}
