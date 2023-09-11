// Super simple file list builder.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	// Version is exported during build
	Version string
)

func main() {
	var err error

	cmd := ""
	url := ""
	flag.StringVar(&cmd, "cmd", "", "Command to execute")
	flag.StringVar(&url, "url", "", "URL to fetch")
	flag.Parse()

	if cmd == "" {
		usage()
		os.Exit(1)
	}
	fmt.Println("Command if true:", cmd)

	if url == "" {
		usage()
		os.Exit(1)
	}
	fmt.Println("URL if false:", url)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("git is not installed, please install git and try again")
		os.Exit(1)
	}

	// get latest head hash
	result, err := run(exec.Command(gitPath, "rev-parse", "-HEAD"))
	if err != nil {
		fmt.Println("Error getting git hash:", err)
		os.Exit(1)
	}
	fmt.Println("Git hash:", result)

	//$(eval NEW_HASH := $(shell git log -n 1 --pretty=format:%H -- charcreate/charcreate.yaml))

	for _, folder := range flag.Args() {
		fmt.Println(folder)
	}
	fmt.Println(gitPath)

}

func usage() {
	fmt.Println("Usage: runifnew -cmd [build] -url [https://download] [folders...]")
	fmt.Println("This program runs build if the git version on folders... matches the latest commit, otherwise it fetches the download url")
}

func run(cmd *exec.Cmd) (string, error) {
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}
