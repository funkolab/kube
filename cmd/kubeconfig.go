package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Kubeconfig struct {
	APIVersion     string      `yaml:"apiVersion"`
	Clusters       []Clusters  `yaml:"clusters"`
	Contexts       []Contexts  `yaml:"contexts"`
	CurrentContext string      `yaml:"current-context"`
	Kind           string      `yaml:"kind"`
	Preferences    Preferences `yaml:"preferences"`
	Users          []Users     `yaml:"users"`
}
type Cluster struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}
type Clusters struct {
	Cluster Cluster `yaml:"cluster"`
	Name    string  `yaml:"name"`
}
type Context struct {
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
}
type Contexts struct {
	Context Context `yaml:"context"`
	Name    string  `yaml:"name"`
}
type Preferences struct {
}
type User struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
}
type Users struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

func (config *Kubeconfig) init(in []byte) {
	err := yaml.Unmarshal(in, &config)
	if err != nil {
		fmt.Println("Kubeconfig format is invalid")
		os.Exit(1)
	}
}
