package main

import (
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/health/", ServeHTTP)
    log.Fatal(http.ListenAndServe(":5000", nil))
}
