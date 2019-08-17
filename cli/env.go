package cli

import (
    "bing/api"
    "fmt"
    "os"
)

func GetBingKey() string {
    bingKey := os.Getenv("BING_API_KEY")
    if bingKey == "" {
        fmt.Println("Cannot run query without a key.")
        os.Exit(1)
    }
    return bingKey
}

func GetBingEndpoint() string {
    bingEndpoint := os.Getenv("BING_ENDPOINT")
    if bingEndpoint == "" {
        bingEndpoint = api.DefaultURL
    }
    return bingEndpoint
}