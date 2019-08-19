package utils

import (
    "io"
    "math/rand"
    "os"
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