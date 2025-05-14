package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ImAyrix/fallparams/runner"
)

func main() {
	var url string
	var output string
	var threads int
	var timeout int
	var verbose bool
	var version bool

	flag.StringVar(&url, "u", "", "")
	flag.StringVar(&output, "o", "", "")
	flag.IntVar(&threads, "t", 25, "")
	flag.IntVar(&timeout, "timeout", 10, "")
	flag.BoolVar(&verbose, "v", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.Parse()

	if version {
		fmt.Println("fallparams version 1.0.0")
		os.Exit(0)
	}

	// Check if URL is provided via stdin
	if url == "" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					url = line
					break // Only take the first URL from stdin
				}
			}
		}
	}

	if url == "" {
		fmt.Println("Usage: fallparams -u <url>")
		fmt.Println("       or")
		fmt.Println("       echo <url> | fallparams")
		os.Exit(1)
	}

	options := &runner.Options{
		Url:     url,
		Output:  output,
		Threads: threads,
		Timeout: timeout,
		Verbose: verbose,
	}

	if err := runner.New(options); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
