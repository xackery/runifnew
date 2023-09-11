// Super simple file list builder.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

var (
	// Version is exported during build
	Version string
)

func main() {
	var err error

	cmd := ""
	url := ""
	urlPath := ""
	flag.StringVar(&cmd, "cmd", "", "Command to execute")
	flag.StringVar(&url, "url", "", "URL to fetch")
	flag.StringVar(&urlPath, "urlPath", "", "Path to save URL to")
	flag.Parse()

	if cmd == "" {
		usage()
		os.Exit(1)
	}

	if url == "" {
		usage()
		os.Exit(1)
	}

	if url != "none" && urlPath == "" {
		usage()
		os.Exit(1)
	}
	fmt.Println("runifnew version:", Version)
	fmt.Println("If TRUE, run:", cmd)
	fmt.Println("If FALSE, fetch:", url, "and save to:", urlPath)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		fmt.Println("git is not installed, please install git and try again")
		os.Exit(1)
	}

	cmdBinary := strings.Split(cmd, " ")[0]
	cmdArgs := strings.Split(cmd, " ")[1:]

	cmdPath, err := exec.LookPath(cmdBinary)
	if err != nil {
		fmt.Println("cmd", cmdBinary, "is not installed, please install cmd and try again")
		os.Exit(1)
	}

	latestHash, err := run(exec.Command(gitPath, "rev-parse", "HEAD"))
	if err != nil {
		fmt.Println("Output:", latestHash)
		fmt.Println("Error getting git hash:", err)
		os.Exit(1)
	}
	if strings.Contains(latestHash, "\n") {
		latestHash = strings.ReplaceAll(latestHash, "\n", "")
	}
	fmt.Println("latest hash:", latestHash)

	isFound := false
	for _, path := range flag.Args() {
		hash, err := run(exec.Command(gitPath, "log", "-n", "1", "--pretty=format:%H", "--", path))
		if err != nil {
			fmt.Println("Output:", hash)
			fmt.Println("Error getting git hash:", err)
			os.Exit(1)
		}
		if strings.Contains(hash, "\n") {
			hash = strings.ReplaceAll(hash, "\n", "")
		}
		fmt.Println(path, "hash:", hash)
		if strings.EqualFold(hash, latestHash) {
			fmt.Println("TRUE:", path, "was updated, triggering cmd")
			isFound = true
			break
		}
	}
	if isFound {
		result, err := run(exec.Command(cmdPath, cmdArgs...))
		if err != nil {
			fmt.Println("Output:", result)
			fmt.Println("Error running cmd:", err)
			os.Exit(1)
		}
		fmt.Println("Output:", result)
		return
	}
	if strings.ToLower(url) == "none" {
		fmt.Println("FALSE: exiting gracefully no url set")
		return
	}
	fmt.Println("FALSE: Fetching url:", url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error fetching url:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println("Error fetching url, status code:", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("Downloaded url:", url)
	f, err := os.Create(urlPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		fmt.Println("Error writing file:", err)
		os.Exit(1)
	}
	fmt.Println("Saved url to:", urlPath)

}

func usage() {
	fmt.Println("Usage: runifnew -cmd [build] -url [https://download] -urlPath bin/test [folders...]")
	fmt.Println("This program runs build if the git version on folders... matches the latest commit, otherwise it fetches the download url. (You can set -url none to do nothing if matches)")
}

func run(cmd *exec.Cmd) (string, error) {
	fmt.Println("Executing command:", strings.Join(cmd.Args, " "))
	cmd.Env = os.Environ()

	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}
