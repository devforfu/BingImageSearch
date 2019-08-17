package io

import (
    "bing/utils"
    "io"
    "net/http"
    "net/url"
    "path"
    "strings"
    "time"
)

type ImageFetcher struct {
    *http.Client
}

var DefaultImageFetcher = NewImageFetcher(10 * time.Second)

func NewImageFetcher(timeout time.Duration) *ImageFetcher {
    return &ImageFetcher{&http.Client{
        CheckRedirect: func(r *http.Request, via []*http.Request) error {
            r.URL.Opaque = r.URL.Path
            return nil
        },
        Timeout: timeout,
    }}
}

func (f *ImageFetcher) Fetch(imageLink, outputFile string) error {
    response, err := f.Get(imageLink)
    if err != nil { return err }
    defer utils.SilentClose(response.Body)

    fileURL, err := url.Parse(imageLink)
    if err != nil { return err }

    segments := strings.Split(fileURL.Path, "/")
    fileName := segments[len(segments) - 1]
    outputFile += path.Ext(fileName)
    file := utils.MustCreateFile(outputFile)
    defer utils.SilentClose(file)

    if _, err = io.Copy(file, response.Body); err != nil {
        return err
    }

    return file.Sync()
}