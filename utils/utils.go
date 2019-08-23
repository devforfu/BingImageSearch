package utils

import (
    "fmt"
    "io"
    "math/rand"
    "net/url"
    "os"
    "path"
    "regexp"
    "strings"
)

func SilentClose(closer io.Closer) {
    _ = closer.Close()
}

func Check(err error) {
    if err != nil {
        panic(err)
    }
}

func MustCreateFile(fileName string) *os.File {
    file, err := os.Create(fileName)
    Check(err)
    return file
}

func SimpleRandomString(size int) string {
    const domain = "abcdef0123456789"
    var b strings.Builder
    for i := 0; i < size; i++ {
        index := rand.Intn(len(domain))
        b.WriteRune(rune(domain[index]))
    }
    return b.String()
}

func FilenameFromURL(fileURL *url.URL) (string, error) {
    segments := strings.Split(fileURL.Path, "/")
    fileName := segments[len(segments) - 1]
    return normalize(path.Ext(fileName))
}

func normalize(extension string) (string, error) {
    patterns := map[string]string {
        "jpg": "\\.(jpeg|JPEG|jpg|JPG).*$",
        "png": "\\.(png|PNG).*$",
    }
    for imageExt, regex := range patterns {
        matched, _ := regexp.MatchString(regex, extension)
        if matched { return imageExt, nil }
    }
    return "", fmt.Errorf("unknown extension: %s", extension)
}