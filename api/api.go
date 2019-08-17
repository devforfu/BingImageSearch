package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
)

const DefaultURL = "https://api.cognitive.microsoft.com/bing/v7.0/images/search"

type SearchParams struct {
	Count int 			`json:"count"`
	Offset int			`json:"offset"`
	Query string		`json:"q"`
	SafeSearch bool		`json:"safeSearch,omitempty"`
	Color string		`json:"color,omitempty"`
	Freshness string 	`json:"freshness,omitempty"`
	ImageType string	`json:"imageType,omitempty"`
	License string		`json:"license,omitempty"`
	Size string			`json:"size,omitempty"`
}
func (p SearchParams) AsQueryParameters() string {
	data, err := json.Marshal(p)
	if err != nil { panic(err) }
	var jsonObject map[string]interface{}
	_ = json.Unmarshal(data, &jsonObject)
	values := make(url.Values)
	for k, v := range jsonObject {
		if v == "" { continue }
		values.Add(k, fmt.Sprintf("%v", v))
	}
	return values.Encode()
}

type ImageResult struct {
	AccentColor string    `json:"accentColor"`
	ContentSize string    `json:"contentSize"`
	EncodingFormat string `json:"encodingFormat"`
	Height int			  `json:"height"`
	Width int			  `json:"width"`
	ImageID string		  `json:"imageId"`
	Name string			  `json:"name"`
	WebSearchURL string   `json:"webSearchUrl"`
	ContentURL string     `json:"contentUrl"`
}

type ImagesCollection struct {
	NextOffset int  	 `json:"nextOffset"`
	Values []ImageResult `json:"value"`
	Query string
}

var DefaultSearchParams = SearchParams{
	Count:150,
	Offset:0,
	Query:"",
	SafeSearch:false,
	Color:"ColorOnly",
	Freshness:"",
	ImageType:"Photo",
	License:"Any",
	Size:"Large",
}

func CreateQuery(query string, offset int) SearchParams {
	params := DefaultSearchParams
	params.Query = query
	params.Offset = offset
	return params
}

type BingClient struct {
	Endpoint string
	SecretKey string
}

func NewBingClient(endpoint, key string) *BingClient {
	return &BingClient{Endpoint:endpoint, SecretKey:key}
}

func (c *BingClient) Pull(queries []string, start int, downloadAll bool) (result []*ImagesCollection) {
	for _, query := range queries {
		if query == "" { continue }
		log.Printf("search string: '%s'", query)
		currOffset := start
		running := true
		for running {
			params := CreateQuery(query, currOffset)
			queryString := params.AsQueryParameters()
			log.Printf("running query with params: %s", queryString)
			images := c.RequestImages(params)
			if downloadAll {
				running = images.NextOffset != currOffset
				currOffset = images.NextOffset
			} else {
				running = false
			}
			result = append(result, images)
		}
	}
	return result
}

func (c *BingClient) PullParallel(queries []string, start int, downloadAll bool) (collections []*ImagesCollection) {
	type result struct {
		collection *ImagesCollection
		err error
	}

	results := make(chan result)
	var wg sync.WaitGroup
	log.Printf("start running queries")

	for i, query := range queries {
		log.Printf("submitting query %d of %d", i, len(queries))

		if query == "" { continue }

		log.Printf("search string: %s", query)

		wg.Add(1)
		go func(q string, output chan<- result) {
			defer wg.Done()
			currOffset := start
			running := true
			var err error
			for running {
				params := CreateQuery(q, currOffset)
				paramsString := params.AsQueryParameters()
				log.Printf("running query with params: %s", paramsString)
				images := c.RequestImages(params)
				if images.Values == nil {
					err = fmt.Errorf("failed to pull query: %s/%s", c.Endpoint, paramsString)
					running = false
				} else if downloadAll {
					running = images.NextOffset != currOffset
					currOffset = images.NextOffset
				} else {
					running = false
				}
				output <- result{images, err}
			}
		}(query, results)
	}

	go func(){
		wg.Wait()
	}()

	log.Printf("combining the collected results")

	for result := range results {
		if result.err != nil {
			log.Printf(result.err.Error())
		} else {
			collections = append(collections, result.collection)
		}
	}

	return collections
}

func (c *BingClient) RequestImages(params SearchParams) *ImagesCollection {
	request := c.MakeRequest("GET", params)
	request.Header.Add("Ocp-Apim-Subscription-Key", c.SecretKey)
	response, err := http.DefaultClient.Do(request)
	if err != nil { panic(err) }
	defer response.Body.Close()
	result := ImagesCollection{}
	decoder := json.NewDecoder(response.Body)
	err = decoder.Decode(&result)
	if err != nil { panic(err) }
	result.Query = params.Query
	return &result
}

func (c *BingClient) MakeRequest(method string, params SearchParams) *http.Request {
	request, err := http.NewRequest(method, c.Endpoint, nil)
	if err != nil { panic(err) }
	query := params.AsQueryParameters()
	request.URL.RawQuery = query
	return request
}
