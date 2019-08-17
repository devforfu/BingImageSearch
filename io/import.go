package io

import (
    "encoding/json"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
)

type Importer func(string, string) ([]string, error)

// FromJSON loads meta information from outputFolder with JSON files.
func FromJSON(outputFolder, fieldName string) ([]string, error) {
    queryLinks := make([]string, 0)
    err := filepath.Walk(outputFolder, func(path string, info os.FileInfo, err error) error {
        if !strings.HasSuffix(path, ".json") { return nil }
        if data, err := ioutil.ReadFile(path); err != nil {
            return err
        } else {
            var content []map[string]interface{}
            err = json.Unmarshal(data, &content)
            if err != nil { return err }
            for _, record := range content {
                url := record[fieldName].(string)
                queryLinks = append(queryLinks, url)
            }
        }
        return nil
    })
    return queryLinks, err
}