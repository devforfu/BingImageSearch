package io

import (
    "bing/utils"
    "fmt"
    "io"
    "net"
    "net/http"
    "net/url"
    "time"
)

type ImageFetcher struct {
    *http.Client
}

func NewImageFetcher(timeout time.Duration) *ImageFetcher {
    var netTransport = &http.Transport{
        Dial: (&net.Dialer{Timeout: timeout}).Dial,
        TLSHandshakeTimeout: timeout,
    }
    return &ImageFetcher{&http.Client{
        CheckRedirect: func(r *http.Request, via []*http.Request) error {
           r.URL.Opaque = r.URL.Path
           return nil
        },
        Transport: netTransport,
        Timeout: timeout,
    }}
}

func (f *ImageFetcher) Fetch(imageLink, outputFile string) (filename string, err error) {
    response, err := f.Get(imageLink)
    if err != nil { return }
    defer utils.SilentClose(response.Body)

    fileURL, err := url.Parse(imageLink)
    if err != nil { return }

    ext, err := utils.FilenameFromURL(fileURL)
    if err != nil { return }

    outputFile += fmt.Sprintf(".%s", ext)
    file := utils.MustCreateFile(outputFile)
    defer utils.SilentClose(file)
    if _, err = io.Copy(file, response.Body); err != nil { return }

    return outputFile, file.Sync()
}