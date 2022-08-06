package cmd

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

func buildList() []kubeChoice {

	var selectList []kubeChoice

	dirname, err := os.UserHomeDir()
	check(err)

	kubieFolder := filepath.Join(dirname, ".kube/kubie")

	files, err := ioutil.ReadDir(kubieFolder)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		fmt.Println("No kubeconfig files found")
		os.Exit(1)
	}

	for _, file := range files {
		if file.Mode().IsRegular() {
			ext := filepath.Ext(file.Name())
			if ext == ".yaml" || ext == ".yml" {

				filePath := filepath.Join(kubieFolder, file.Name())
				config := clientcmd.GetConfigFromFileOrDie(filePath)

				for name, context := range config.Contexts {

					choice := kubeChoice{Name: name, ContextName: name, Context: context, Config: config, filePath: filePath}

					if flag.NArg() == 1 && !strings.Contains(name, flag.Arg(0)) {
						continue
					}
					if isTokenExpired(choice) {
						continue
					}
					selectList = append(selectList, choice)
				}

			}
		}
	}
	return selectList
}

func isTokenExpired(choice kubeChoice) bool {

	tokenString := choice.Config.AuthInfos[choice.Context.AuthInfo].Token
	if tokenString != "" {
		claims := jwt.MapClaims{}
		token, _ := jwt.ParseWithClaims(tokenString, claims, nil)

		if token != nil {
			exp := int64(claims["exp"].(float64))
			if exp < time.Now().Unix() {
				if len(choice.Config.Contexts) == 1 {
					os.Remove(choice.filePath)
				}
				return true
			}
		}

	}
	return false
}

func checkList(selectList *[]kubeChoice) {
	if len(*selectList) == 0 {
		if flag.NArg() == 1 {
			fmt.Printf("No cluster found with the filter \"%s\"\n", flag.Arg(0))
		} else {
			fmt.Println("No clusters found")
		}
		os.Exit(1)
	}
}
