package main

import (
    "bufio"
    "github.com/ImAyrix/fallparams/funcs/opt"
    "github.com/ImAyrix/fallparams/funcs/run"
    "github.com/ImAyrix/fallparams/funcs/validate"
    "github.com/projectdiscovery/goflags"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "sync"
)

var (
    wg        sync.WaitGroup
    myOptions = &opt.Options{}
)

func ReadFlags() *goflags.FlagSet {
    flagSet := goflags.NewFlagSet()
    flagSet.SetDescription("Find All Parameters")

    flagSet.StringVarP(&myOptions.InputUrls, "url", "u", "", "Input [Filename | URL]")
    flagSet.StringVarP(&myOptions.InputDIR, "directory", "dir", "", "Stored requests/responses files directory path (offline)")
    flagSet.IntVarP(&myOptions.Thread, "thread", "t", 1, "Number of Threads [Number]")
    flagSet.IntVarP(&myOptions.Delay, "delay", "rd", 0, "Request delay between each request in seconds")
    flagSet.StringVarP(&myOptions.OutputFile, "output", "o", "parameters.txt", "File to write output to")
    flagSet.IntVarP(&myOptions.MaxLength, "max-length", "xl", 30, "Maximum length of words")
    flagSet.IntVarP(&myOptions.MinLength, "min-length", "nl", 0, "Minimum length of words")
    flagSet.BoolVarP(&myOptions.CrawlMode, "crawl", "c", false, "Crawl pages to extract their parameters")
    flagSet.BoolVarP(&myOptions.Headless, "headless", "hl", false, "Discover parameters with headless browser")
    flagSet.BoolVarP(&myOptions.DisableUpdateCheck, "disable-update-check", "duc", false, "Disable automatic fallparams update check")
    flagSet.IntVarP(&myOptions.MaxDepth, "depth", "d", 2, "maximum depth to crawl")
    flagSet.VarP(&myOptions.CustomHeaders, "header", "H", "Header `\"Name: Value\"`, separated by colon. Multiple -H flags are accepted.")
    flagSet.StringVarP(&myOptions.RequestHttpMethod, "method", "X", "GET", "HTTP method to use")
    flagSet.StringVarP(&myOptions.RequestBody, "body", "b", "", "POST data")
    flagSet.StringVarP(&myOptions.InputHttpRequest, "request", "r", "", "File containing the raw http request")
    flagSet.StringVarP(&myOptions.ProxyUrl, "proxy", "x", "", "Proxy URL (SOCKS5 or HTTP)")
    flagSet.BoolVar(&myOptions.SilentMode, "silent", false, "Disables the banner and prints output to the command line")

    err := flagSet.Parse()
    if err != nil {
        os.Exit(1)
    }

    // Check for stdin input if -u is not provided
    if myOptions.InputUrls == "" {
        stat, _ := os.Stdin.Stat()
        if (stat.Mode() & os.ModeCharDevice) == 0 { // Check if data is piped
            scanner := bufio.NewScanner(os.Stdin)
            var urls []string
            for scanner.Scan() {
                url := strings.TrimSpace(scanner.Text())
                if url != "" && validate.IsUrl(url) && !strings.Contains(url, "{==MY=FILE=NAME==}") {
                    urls = append(urls, url)
                }
            }
            if len(urls) > 0 {
                myOptions.InputUrls = strings.Join(urls, ",")
            }
        }
    }

    return flagSet
}

func main() {
    ReadFlags()

    urlChannel := make(chan string, myOptions.Thread)

    for i := 0; i < myOptions.Thread; i++ {
        wg.Add(1)
        go run.Start(urlChannel, myOptions, &wg)
    }

    // Process InputUrls (-u)
    if myOptions.InputUrls != "" {
        // Check if InputUrls is a file
        if _, err := os.Stat(myOptions.InputUrls); err == nil {
            data, err := ioutil.ReadFile(myOptions.InputUrls)
            if err != nil {
                println("Error reading file:", err.Error())
                os.Exit(1)
            }
            urls := strings.Split(string(data), "\n")
            for _, url := range urls {
                url = strings.TrimSpace(url)
                if url != "" && validate.IsUrl(url) && !strings.Contains(url, "{==MY=FILE=NAME==}") {
                    urlChannel <- url
                }
            }
        } else {
            urls := strings.Split(myOptions.InputUrls, ",")
            for _, url := range urls {
                url = strings.TrimSpace(url)
                if url != "" && validate.IsUrl(url) && !strings.Contains(url, "{==MY=FILE=NAME==}") {
                    urlChannel <- url
                }
            }
        }
    }

    // Process InputHttpRequest (-r)
    if myOptions.InputHttpRequest != "" {
        data, err := ioutil.ReadFile(myOptions.InputHttpRequest)
        if err != nil {
            println("Error reading request file:", err.Error())
            os.Exit(1)
        }
        urlChannel <- string(data)
    }

    // Process InputDIR (-dir)
    if myOptions.InputDIR != "" {
        files, err := ioutil.ReadDir(myOptions.InputDIR)
        if err != nil {
            println("Error reading directory:", err.Error())
            os.Exit(1)
        }
        for _, file := range files {
            if !file.IsDir() {
                data, err := ioutil.ReadFile(filepath.Join(myOptions.InputDIR, file.Name()))
                if err != nil {
                    continue
                }
                input := file.Name() + "{==MY=FILE=NAME==}" + string(data)
                if strings.Contains(input, "{==MY=FILE=NAME==}") {
                    urlChannel <- input
                }
            }
        }
    }

    close(urlChannel)
    wg.Wait()
}
