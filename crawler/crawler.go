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
    client := c.Client
    resultsQueue := make(chan result, 10)
    queriesQueue := make(chan string)
    utils.Check(os.MkdirAll(outputFolder, os.ModePerm))

    go func(){
        for _, query := range queries {
            queriesQueue <- query
        }
        close(queriesQueue)
    }()

    var workerGroup sync.WaitGroup
    for i := 1; i <= c.NumWorkers; i++ {
        log.Printf("submitting querying worker %d of %d", i, c.NumWorkers)
        workerGroup.Add(1)
        go queryWorker(i, queriesQueue, resultsQueue, &workerGroup, client)
    }

    go func() {
        log.Printf("wait for workers to close the processing queue")
        workerGroup.Wait()
        log.Printf("closing the queue")
        close(resultsQueue)
    }()

    var writerGroup sync.WaitGroup
    for i := 1; i <= c.NumWorkers; i++ {
        log.Printf("Submitting writing worker %d of %d", i, c.NumWorkers)
        writerGroup.Add(1)
        go writingWorker(i, outputFolder, resultsQueue, &writerGroup, exportFunc)
    }

    log.Printf("waiting for writers...")
    writerGroup.Wait()
    log.Printf("collected results are saved into folder: %s", outputFolder)
}

type result struct {
    collection *api.ImagesCollection
    err error
}

func queryWorker(
    workerIndex int,
    in <-chan string,
    out chan<- result,
    group *sync.WaitGroup,
    client *api.BingClient) {

    defer group.Done()

    for queryString := range in {
        if queryString == "" { continue }
        log.Printf("[worker:%d] sending search string: %s", workerIndex, queryString)

        currOffset := 0
        running := true
        var err error
        for running {
            params := api.CreateQuery(queryString, currOffset)
            paramsString := params.AsQueryParameters()
            log.Printf("[worker:%d] running query with params: %s", workerIndex, paramsString)
            images := client.RequestImages(params)
            if images.Values == nil {
                err = fmt.Errorf("[worker:%d] failed to pull query: %s/%s",
                    workerIndex, client.Endpoint, paramsString)
                running = false
            } else {
                running = images.NextOffset != currOffset
                currOffset = images.NextOffset
            }
            out <- result{images, err}
        }
    }

    log.Printf("[worker:%d] terminated", workerIndex)
}

func writingWorker(
    workerIndex int,
    outputFolder string,
    in <-chan result,
    group *sync.WaitGroup,
    exportFunc io.Exporter) {

    defer group.Done()

    for result := range in {
        outputFile := path.Join(outputFolder, utils.RandomString(20))
        log.Printf("[worker:%d] saving file %s", workerIndex, outputFile)
        if result.err != nil {
            log.Printf(result.err.Error())
        } else {
            query := result.collection.Query
            log.Printf(
                "[worker:%d] exporting query results for '%s' into file '%s",
                workerIndex, query, outputFile)
            if err := exportFunc(result.collection, outputFile); err != nil {
                log.Printf(err.Error())
            }
        }
    }

    log.Printf("[worker:%d] terminated", workerIndex)
}

func (c *Crawler) Download(metaDataFolder, imagesFolder string, importFunc io.Importer) {
    log.Printf("loading image URLs from folder: %s", metaDataFolder)

    utils.Check(os.MkdirAll(imagesFolder, os.ModePerm))
    urls, err := importFunc(metaDataFolder, "contentUrl")
    if err != nil { log.Printf("%s", err) }

    feed := make(chan string, c.NumWorkers)
    go func () {
        for _, url := range urls {
            // log.Printf("submitting URL: %s", url)
            feed <- url
        }
        close(feed)
    }()

    log.Printf("launching workers...")
    var workerGroup sync.WaitGroup
    results := make(chan downloaded)
    for i := 1; i <= c.NumWorkers; i++ {
        workerGroup.Add(1)
        go downloader(i, imagesFolder, feed, results, &workerGroup)
    }

    go func(){
        workerGroup.Wait()
        log.Printf("all downloaders were terminated, closing results channel")
        close(results)
    }()

    collected := make(map[string]bool)
    for result := range results {
        collected[result.url] = result.success
    }

    collectedJSON, _ := json.Marshal(collected)
    metaFile := path.Join(imagesFolder, "collected.json")
    _ = ioutil.WriteFile(metaFile, collectedJSON, os.ModePerm)
    log.Printf("collected results are saved into folder: %s", imagesFolder)
}

type downloaded struct {
    url string
    success bool
}

func downloader(
    workerIndex int,
    imagesFolder string,
    urls <-chan string,
    results chan<- downloaded,
    group *sync.WaitGroup) {

    defer group.Done()

    fetcher := io.NewImageFetcher(1*time.Hour)
    for url := range urls {
        log.Printf("[worker:%d] fetching URL: %s", workerIndex, url)
        outputFile := path.Join(imagesFolder, utils.RandomString(20))
        success := true
        err := fetcher.Fetch(url, outputFile)
        if err != nil {
            log.Printf("[worker:%d] %s", workerIndex, err.Error())
            success = false
        }
        results <- downloaded{url, success}
    }

    log.Printf("[worker:%d] terminated", workerIndex)
}