package export

import (
    "bing/api"
    "fmt"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
)

var DefaultHeader = []string {
    "Query",
    "Name",
    "AccentColor",
    "ContentSize",
    "Width",
    "Height",
    "Format",
    "URL",
}

// ToCSV saves the collected information about images to use it
// later for downloading.
func ToCSV(collection *api.ImagesCollection, outputFile string) {

    if _, err := os.Stat(outputFile); os.IsNotExist(err) {
        header := strings.Join(DefaultHeader, ",")
        err := ioutil.WriteFile(outputFile, []byte(header), os.ModePerm)
        if err != nil { panic(err) }
    }

    f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_APPEND, os.ModePerm)
    if err != nil { panic(err) }

    for _, meta := range collection.Values {
        var lineItems []string
        lineItems = append(lineItems, collection.Query)
        lineItems = append(lineItems, meta.Name)
        lineItems = append(lineItems, meta.AccentColor)
        lineItems = append(lineItems, meta.ContentSize)
        lineItems = append(lineItems, strconv.Itoa(meta.Width))
        lineItems = append(lineItems, strconv.Itoa(meta.Height))
        lineItems = append(lineItems, meta.EncodingFormat)
        lineItems = append(lineItems, meta.ContentURL)
        line := strings.Join(lineItems, ",") + "\n"
        _, err = f.WriteString(line)
        if err != nil {
            fmt.Printf("Warning: %s\n", err.Error())
        } else {
            _ = f.Sync()
        }
    }
    if err := f.Close(); err != nil {
        panic(err)
    }
}