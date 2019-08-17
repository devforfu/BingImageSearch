package cli

import (
    "flag"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

type RunConfig struct {
    DownloadAll *bool
    Offset *int
    Query *string
    File *string
    OutputFolder *string
    QueryList []string
}

func ParseArguments() *RunConfig {
    conf := RunConfig{}
    conf.Query = flag.String("q", "", "search query")
    conf.File = flag.String("f", "", "a path to the file with search queries, one per line")
    conf.DownloadAll = flag.Bool("a", false, "download all query results, not only the offset page")
    conf.Offset = flag.Int("p", 0, "take results starting with offset")
    conf.OutputFolder = flag.String("o", "output", "path to the folder with dumped queries")
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
