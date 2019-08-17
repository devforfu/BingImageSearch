package main

import (
    "bing/api"
    "bing/cli"
    "bing/crawler"
    "bing/io"
)

func main() {
    conf := cli.ParseArguments()
    client := api.NewBingClient(cli.GetBingEndpoint(), cli.GetBingKey())
    crawl := crawler.Crawler{Client:client, NumWorkers:10}
    switch *conf.Mode {
    case "query": crawl.Crawl(conf.QueryList, *conf.OutputFolder, io.ToJSON)
    case "download": crawl.Download(*conf.File, *conf.OutputFolder, io.FromJSON)
    }
}