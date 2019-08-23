package crawler

import (
    "bing/api"
    "bing/io"
    "bing/utils"
    "fmt"
    "log"
    "path"
    "sync"
    "time"
)

// queryWorker performs REST API queries taking search strings from in channel, and saving
// results into out channel.
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

// writingWorker takes JSON structures from in channel and saves them onto disk.
func writingWorker(
    workerIndex int,
    outputFolder string,
    in <-chan result,
    group *sync.WaitGroup,
    exportFunc io.Exporter) {

    defer group.Done()

    for result := range in {
        outputFile := path.Join(outputFolder, utils.SimpleRandomString(20))
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

// downloadingWorker performs actual work of retrieving the images and saving them onto local disk.
func downloadingWorker(
    workerIndex int,
    imagesFolder string,
    urls <-chan string,
    results chan<- Downloaded,
    group *sync.WaitGroup) {

    defer group.Done()

    fetcher := io.NewImageFetcher(1*time.Hour)
    for url := range urls {
        log.Printf("[worker:%d] fetching URL: %s", workerIndex, url)
        outputFile := path.Join(imagesFolder, utils.SimpleRandomString(20))
        createdFile, err := fetcher.Fetch(url, outputFile)
        if err != nil {
            log.Printf("[worker:%d] %s", workerIndex, err.Error())
        }
        results <- Downloaded{URL:url, Filename:createdFile, Error:err}
    }

    log.Printf("[worker:%d] terminated", workerIndex)
}

// enqueueStrings sends strings into channel, and closes it.
func enqueueStrings(strings []string, channel chan<- string) {
    for _, item := range strings {
        channel <- item
    }
    close(channel)
}
