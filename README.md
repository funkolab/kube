## kube - Quickly connect to your Kubernetes cluster!

CLI tools to manage your kubeconfig file


---

[![Go Report Card](https://goreportcard.com/badge/github.com/funkolab/kube)](https://goreportcard.com/report/github.com/funkolab/kube)
[![Maintainability](https://api.codeclimate.com/v1/badges/e0f43f5c74eabfa8bc4d/maintainability)](https://codeclimate.com/github/funkolab/kube/maintainability)
![CI](https://github.com/funkolab/kube/actions/workflows/test.yaml/badge.svg)
![Release](https://github.com/funkolab/kube/actions/workflows/release.yaml/badge.svg)

[![release](https://img.shields.io/github/release-pre/funkolab/kube.svg)](https://github.com/funkolab/kube/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/funkolab/kube/blob/master/LICENSE)
![Proudly written in Golang](https://img.shields.io/badge/written%20in-Golang-92d1e7.svg)
[![Releases](https://img.shields.io/github/downloads/funkolab/kube/total.svg)](https://github.com/funkolab/kube/releases)

---



## Installation

There are several installation options:

- As Homebrew or Linuxbrew package
- Manual installation

After installing, the tools will be available as `kube`.

### Homebrew Package

You can install with [Homebrew](https://brew.sh) for macOS or [LinuxBrew](https://docs.brew.sh/Homebrew-on-Linux) for Linux

```sh
brew install funkolab/tap/kube
```

Keep up-to-date with `brew upgrade kube` (or brew upgrade to upgrade everything)

### Manual

 - Download your corresponding [release](https://github.com/funkolab/kube/releases)
 - Install the binary somewhere in your PATH (/usr/local/bin for example)
 - use it with `kube`

***MacOS X notes for security error***

 Depending of your OS settings when you install you binary manually we must launch the following command:
 `xattr -r -d com.apple.quarantine /usr/local/bin/kube`

## Usage

TODO



## Building From Source

 kube is currently using go v1.17 or above. In order to build  kube from source you must:

 1. Clone the repo
 2. Build and run the executable

      ```shell
      make build && make install
      ```