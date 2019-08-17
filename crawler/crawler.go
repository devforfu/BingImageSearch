package crawler

import (
    "bing/api"
    "bing/io"
    "bing/utils"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path"
    "sync"
    "time"
)

type Crawler struct {
    Client *api.BingClient
    NumWorkers int
}

func (c *Crawler) Crawl(queries []string, outputFolder string, exportFunc io.Exporter) {
    type result struct {
        collection *api.ImagesCollection
        err error
    }

    client := c.Client
    queue := make(chan result, 10)

    var workGroup sync.WaitGroup
    log.Printf("launch workers")


    for i, query := range queries {
        log.Printf("submitting query %d of %d", i+1, len(queries))

        if query == "" { continue }

        workGroup.Add(1)
        go func(q string, out chan<- result) {
            defer workGroup.Done()
            log.Printf("search string: %s", q)
            currOffset := 0
            running := true
            var err error
            for running {
                params := api.CreateQuery(q, currOffset)
                paramsString := params.AsQueryParameters()
                log.Printf("running query with params: %s", paramsString)
                images := client.RequestImages(params)
                if images.Values == nil {
                    err = fmt.Errorf("failed to pull query: %s/%s", client.Endpoint, paramsString)
                    running = false
                } else {
                    running = images.NextOffset != currOffset
                    currOffset = images.NextOffset
                }
                queue <- result{images, err}
            }
        }(query, queue)
    }

    go func(){
        log.Printf("wait for workers to close the processing queue")
        workGroup.Wait()
        log.Printf("closing the queue")
        close(queue)
    }()

    if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
        log.Fatalf("failed to create output folder: %s", err.Error())
    }

    log.Printf("launch writers...")

    var writerGroup sync.WaitGroup

    for r := range queue {
        writerGroup.Add(1)
        go func(r result) {
            defer writerGroup.Done()
            outputFile := path.Join(outputFolder, utils.RandomString(20))
            if r.err != nil {
                log.Printf("%s", r.err.Error())
            } else {
                query := r.collection.Query
                log.Printf("exporting query results for '%s' into file %s", query, outputFile)
                err := exportFunc(r.collection, outputFile)
                if err != nil {
                    log.Printf(err.Error())
                }
            }
        }(r)
    }

    log.Println("waiting for writers...")
    writerGroup.Wait()
}

func (c *Crawler) Download(metaDataFolder, imagesFolder string, importFunc io.Importer) {
    log.Printf("loading image URLs from folder: %s", imagesFolder)

    utils.Check(os.MkdirAll(imagesFolder, os.ModePerm))
    urls, err := importFunc(metaDataFolder, "contentUrl")
    if err != nil { log.Printf("%s", err) }

    log.Printf("launching workers...")
    var wg sync.WaitGroup
    fetcher := io.NewImageFetcher(1*time.Hour)

    collected := make(map[string]string)
    for _, url := range urls {
        log.Printf("fetching URL: %s", url)
        wg.Add(1)
        go func() {
            defer wg.Done()
            outputFile := path.Join(imagesFolder, utils.RandomString(20))
            err := fetcher.Fetch(url, outputFile)
            if err != nil {
               log.Printf(err.Error())
            } else {
                collected[outputFile] = url
            }
        }()
    }

    log.Printf("waiting for data fetchers...")
    wg.Wait()

    log.Printf("writing meta data")
    collectedJson, _ := json.Marshal(collected)
    metaFile := path.Join(imagesFolder, "collected.json")
    _ = ioutil.WriteFile(metaFile, collectedJson, os.ModePerm)
}