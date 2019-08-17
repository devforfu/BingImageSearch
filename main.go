package main

import (
    "bing/api"
    "bing/export"
    "flag"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

func main() {
    conf := parseArguments()
    client := api.NewBingClient(getBingEndpoint(), getBingKey())
    result := client.Pull(conf.QueryList, *conf.Offset, *conf.DownloadAll)
    for _, batch := range result {
        export.ToCSV(batch, *conf.OutputPath)
    }
}

type RunConfig struct {
    DownloadAll *bool
    Offset *int
    Query *string
    File *string
    OutputPath *string
    QueryList []string
}

func parseArguments() *RunConfig {
    conf := RunConfig{}
    conf.Query = flag.String("q", "", "search query")
    conf.File = flag.String("f", "", "a path to the file with search queries, one per line")
    conf.DownloadAll = flag.Bool("a", false, "download all query results, not only the offset page")
    conf.Offset = flag.Int("p", 0, "take results starting with offset")
    conf.OutputPath = flag.String("o", "output.csv", "path to dump queries")
    flag.Parse()

    if (*conf.Query == "") && (*conf.File == "") {
        log.Fatalln("Cannot run search without -q or -f arguments provided.")
    } else if (*conf.Query != "") && (*conf.File != "") {
        log.Fatalln("Ambiguous arguments: both -q and -f are specified.")
    }

    fileName := *conf.File
    useFile := fileName != ""
    if useFile {
        if data, err := ioutil.ReadFile(fileName); err != nil {
            if os.IsNotExist(err) { log.Fatalf("File doesn't exist: %s", fileName) }
            if os.IsPermission(err) { log.Fatalf("Permission error: %s", err.Error()) }
        } else {
            content := string(data)
            conf.QueryList = strings.Split(content, "\n")
        }
    } else {
        conf.QueryList = []string{*conf.Query}
    }

    return &conf
}

func getBingKey() string {
    bingKey := os.Getenv("BING_API_KEY")
    if bingKey == "" {
        fmt.Println("Cannot run query without a key.")
        os.Exit(1)
    }
    return bingKey
}

func getBingEndpoint() string {
    bingEndpoint := os.Getenv("BING_ENDPOINT")
    if bingEndpoint == "" {
        bingEndpoint = api.DefaultURL
    }
    return bingEndpoint
}