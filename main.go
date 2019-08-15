package main

import (
    "bing/api"
    "fmt"
    "os"
)

func main() {
    bingKey := os.Getenv("BING_API_KEY")
    if bingKey == "" {
        fmt.Printf("Cannot run query without a key")
        os.Exit(1)
    }
    params := api.CreateQuery("dog", 1)
    client := api.BingClient{SecretKey:bingKey}
    images := client.RequestImages(params)
    fmt.Print(images)
}
