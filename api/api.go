package api

import (
	"encoding/json"
	"net/http"
)

const baseURL = "https://api.cognitive.microsoft.com/bing/v7.0/images/search"

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

type ImageResult struct {
	AccentColor string    `json:"accentColor"`
	ContentSize string    `json:"contentSize"`
	EncodingFormat string `json:"encodingFormat"`
	Height int			  `json:"height"`
	Width int			  `json:"width"`
	ImageID string		  `json:"imageId"`
	Name string			  `json:"name"`
	WebSearchUrl string   `json:"webSearchUrl"`
}

type ImagesCollection struct {
	ID string 				 `json:"id"`
	NextOffset int  		 `json:"nextOffset"`
	Collection []ImageResult `json:"value"`
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
	SecretKey string
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
	return &result
}

func (c *BingClient) MakeRequest(method string, params SearchParams) *http.Request {
	request, err := http.NewRequest(method, baseURL, nil)
	if err != nil { panic(err) }
	data, err := json.Marshal(params)
	if err != nil { panic(err) }
	var jsonObject map[string]string
	_ = json.Unmarshal(data, &jsonObject)
	q := request.URL.Query()
	for k, v := range jsonObject {
		if v == "" { continue }
		q.Add(k, v)
	}
	request.URL.RawQuery = q.Encode()
	return request
}
