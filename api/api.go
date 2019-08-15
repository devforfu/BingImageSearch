package api

import "net/http"

const baseURL = "https://api.cognitive.microsoft.com/bing/v7.0/images/search"

type ColorOption int

type SearchParams struct {
	Count int
	Offset int
	Query string
	SafeSearch bool
}

type BingClient struct {
	SecretKey string
}

func createHeaders(key string) http.Header {
	return http.Header{"Ocp-Apim-Subscription-Key" : []string{key}}
}