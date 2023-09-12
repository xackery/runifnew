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
	Version   string
	isVerbose bool
)

func main() {
	var err error

	if len(os.Args) > 0 && os.Args[1] == "version" {
		fmt.Println("runifnew version:", Version)
		os.Exit(0)
	}

	cmd := ""
	url := ""
	urlPath := ""
	flag.StringVar(&cmd, "cmd", "", "Command to execute")
	flag.StringVar(&url, "url", "", "URL to fetch")
	isVerbose = *flag.Bool("v", false, "Verbose output")
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
	println("If TRUE, run:", cmd)
	println("If FALSE, fetch:", url, "and save to:", urlPath)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		println("git is not installed, please install git and try again")
		os.Exit(1)
	}

	cmdBinary := strings.Split(cmd, " ")[0]
	cmdArgs := strings.Split(cmd, " ")[1:]

	cmdPath, err := exec.LookPath(cmdBinary)
	if err != nil {
		println("cmd", cmdBinary, "is not installed, please install cmd and try again")
		os.Exit(1)
	}

	latestHash, err := run(exec.Command(gitPath, "rev-parse", "HEAD"))
	if err != nil {
		println("Output:", latestHash)
		println("Error getting git hash:", err)
		os.Exit(1)
	}
	if strings.Contains(latestHash, "\n") {
		latestHash = strings.ReplaceAll(latestHash, "\n", "")
	}
	println("latest hash:", latestHash)

	isFound := false
	for _, path := range flag.Args() {
		hash, err := run(exec.Command(gitPath, "log", "-n", "1", "--pretty=format:%H", "--", path))
		if err != nil {
			println("Output:", hash)
			println("Error getting git hash:", err)
			os.Exit(1)
		}
		if strings.Contains(hash, "\n") {
			hash = strings.ReplaceAll(hash, "\n", "")
		}
		println(path, "hash:", hash)
		if strings.EqualFold(hash, latestHash) {
			println("TRUE:", path, "was updated, triggering cmd")
			isFound = true
			break
		}
	}
	if isFound {
		fmt.Println("[runifnew] Running cmd:", cmd)
		result, err := run(exec.Command(cmdPath, cmdArgs...))
		if err != nil {
			println("Output:", result)
			println("Error running cmd:", err)
			os.Exit(1)
		}
		println("Output:", result)
		return
	}
	if strings.ToLower(url) == "none" {
		println("FALSE: exiting gracefully no url set")
		return
	}
	println("FALSE: Fetching url:", url)
	resp, err := http.Get(url)
	if err != nil {
		println("Error fetching url:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		println("Error fetching url, status code:", resp.StatusCode)
		os.Exit(1)
	}
	println("Downloaded url:", url)
	f, err := os.Create(urlPath)
	if err != nil {
		println("Error creating file:", err)
		os.Exit(1)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		println("Error writing file:", err)
		os.Exit(1)
	}
	fmt.Println("[runifnew] Saved url to:", urlPath)

}

func usage() {
	fmt.Println("Usage: runifnew -cmd [build] -url [https://download] -urlPath bin/test [folders...]")
	fmt.Println("This program runs build if the git version on folders... matches the latest commit, otherwise it fetches the download url. (You can set -url none to do nothing if matches)")
}

func run(cmd *exec.Cmd) (string, error) {
	println("Executing command:", strings.Join(cmd.Args, " "))
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

func println(a ...interface{}) {
	if isVerbose {
		fmt.Printf("[runifnew] ")
		fmt.Println(a...)
	}
}
