package crawler

import (
    "bing/api"
    "bing/export"
    "fmt"
    "log"
    "math/rand"
    "os"
    "path"
    "strings"
    "sync"
)

type Crawler struct {
    Client *api.BingClient
}

func (c *Crawler) Crawl(queries []string, outputFolder string) {
    type result struct {
        collection *api.ImagesCollection
        err error
    }

    client := c.Client
    queue := make(chan result, 10)

    var workGroup sync.WaitGroup
    log.Printf("start running queries")

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

    const domain = "abcdef0123456789"
    var randomString = func(size int) string {
        var b strings.Builder
        for i := 0; i < size; i++ {
            index := rand.Intn(len(domain))
            b.WriteRune(rune(domain[index]))
        }
        return b.String()
    }

    if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
        log.Fatalf("failed to create output folder: %s", err.Error())
    }

    log.Printf("launch writers...")

    var writerGroup sync.WaitGroup
    writerGroup.Add(1)
    for r := range queue {
        go func(r result){
            defer writerGroup.Done()
            outputFile := path.Join(outputFolder, randomString(20) + ".csv")
            if r.err != nil {
                log.Printf("%s", r.err.Error())
            } else {
                query := r.collection.Query
                log.Printf("exporting query results for '%s' into file %s", query, outputFile)
                export.ToCSV(r.collection, outputFile)
            }
        }(r)
    }

    log.Println("waiting for writers...")
    writerGroup.Wait()
}