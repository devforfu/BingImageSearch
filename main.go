package main

import (
    "bing/api"
    "bing/cli"
    "bing/crawler"
)

func main() {
    conf := cli.ParseArguments()
    client := api.NewBingClient(cli.GetBingEndpoint(), cli.GetBingKey())
    crawl := crawler.Crawler{Client:client}
    crawl.Crawl(conf.QueryList, *conf.OutputFolder)
}