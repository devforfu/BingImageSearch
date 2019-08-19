package crawler

import (
    "bing/api"
    "bing/io"
    "bing/utils"
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "path"
    "sync"
)

// Crawler takes Bing API client instance and sends queries to the search engine.
// It also responsible for taking URLs from search results and using them to
// download the images into a local folder.
type Crawler struct {
    Client *api.BingClient
    NumWorkers int
}

// result contains a collection of URLs from query, or error if query failed.
type result struct {
    collection *api.ImagesCollection
    err error
}

// Crawl takes list of strings and send them (in parallel) the images search endpoint.
// A result of each query represents a JSON object that is saved onto local disk with
// exportFunc into outputFolder.
func (c *Crawler) Crawl(queries []string, outputFolder string, exportFunc io.Exporter) {
    client := c.Client
    resultsQueue := make(chan result, 10)
    queriesQueue := make(chan string)
    utils.Check(os.MkdirAll(outputFolder, os.ModePerm))

    go enqueueStrings(queries, queriesQueue)

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

// downloaded contains image URL and downloading success status.
type downloaded struct {
    url string
    success bool
}

// Download takes previously retrieved queries results from metaDataFolder and starts
// downloading them into imagesFolder. The importFunc is used to read queries files
// from disk.
func (c *Crawler) Download(metaDataFolder, imagesFolder string, importFunc io.Importer) {
    log.Printf("loading image URLs from folder: %s", metaDataFolder)

    utils.Check(os.MkdirAll(imagesFolder, os.ModePerm))
    urls, err := importFunc(metaDataFolder, "contentUrl")
    if err != nil{
        log.Printf("%s", err)
        return
    }

    feed := make(chan string, c.NumWorkers)
    go enqueueStrings(urls, feed)

    log.Printf("launching workers...")
    var workerGroup sync.WaitGroup
    results := make(chan downloaded)
    for i := 1; i <= c.NumWorkers; i++ {
        workerGroup.Add(1)
        go downloadingWorker(i, imagesFolder, feed, results, &workerGroup)
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
