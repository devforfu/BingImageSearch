package io

import (
    "bing/api"
    "encoding/json"
    "io/ioutil"
    "os"
    "strconv"
    "strings"
)

var DefaultHeaderCSV = []string {
    "Query",
    "Name",
    "AccentColor",
    "ContentSize",
    "Width",
    "Height",
    "Format",
    "URL",
}

type Exporter func(*api.ImagesCollection, string) error

// ToCSV saves the collected information about images to use it
// later for downloading.
func ToCSV(collection *api.ImagesCollection, outputFile string) error {
    outputFile += ".csv"

    if _, err := os.Stat(outputFile); os.IsNotExist(err) {
        header := strings.Join(DefaultHeaderCSV, ",")
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
            return err
        } else if err = f.Sync(); err != nil {
            return err
        }
    }

    if err := f.Close(); err != nil {
        panic(err)
    }

    return nil
}

func ToJSON(collection *api.ImagesCollection, outputFile string) error {
    outputFile += ".json"

    f, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
    if err != nil { panic(err) }

    defer func(){
       if fileErr := f.Close(); fileErr != nil {
           panic(fileErr)
       }
    }()

    if serialized, err := json.MarshalIndent(collection.Values, "", " "); err != nil {
        return err
    } else if _, err = f.Write(serialized); err != nil {
        return err
    }

    return f.Sync()
}