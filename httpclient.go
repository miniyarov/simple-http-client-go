// Author: Ulugbek Miniyarov
package main

import (
    "os"
    "io"
    "io/ioutil"
    "fmt"
    "net"
    "net/http"
    "compress/gzip"
    "time"
    "encoding/json"
    "strings"
    "bytes"
    "sync"
)

type RequestSchema struct {
    Method string         `json:"method"`
    Url string             `json:"url"`
    Headers interface{}    `json:"headers"`
    Body string            `json:"body"`
}

func Client(method string, url string, headers interface{}, body io.Reader) (resp *http.Response, err error) {
    // Override default (30 Second) connection/dial timeout
    // Ref: https://golang.org/src/net/http/transport.go
    netTransport := &http.Transport{
        DialContext: (&net.Dialer{
            Timeout: 3 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout: 3 * time.Second,
    }

    // Override default (0 Second / indefinite) response timeout
    // Ref: https://golang.org/src/net/http/client.go
    netClient := &http.Client{
        Timeout: 20 * time.Second,
        Transport: netTransport,
    }

    // Prepares request object
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, err
    }

    // Adds given headers to request object
    for header, value := range headers.(map[string]interface{}) {
        req.Header.Set(header, value.(string))
    }

    // Sends request and returns response object
    return netClient.Do(req)
}

func Request(method string, url string, headers interface{}, body string) {
    // Execute request
    resp, err := Client(method, url, headers, bytes.NewBuffer([]byte(body)))
    if err != nil {
        // log request error
        panic(err)
    }
    defer resp.Body.Close()

    var responseBody io.ReadCloser
    switch resp.Header.Get("Content-Encoding") {
    case "gzip":
        responseBody, err = gzip.NewReader(resp.Body)
        if err != nil {
            panic(err)
        }
        defer responseBody.Close()
    default:
        responseBody = resp.Body
    }

    responseBytes, err := ioutil.ReadAll(responseBody)
    if err != nil {
        // log response read error
        panic(err)
    }

    // output response
    fmt.Println(string(responseBytes))
}

func parsePayload(payload string) []RequestSchema {
    requests := make([]RequestSchema, 0, 20)

    decoder := json.NewDecoder(strings.NewReader(payload))
    _, err := decoder.Token()
    if err != nil {
        panic(err)
    }

    var request RequestSchema
    for decoder.More() {
        err := decoder.Decode(&request)
        if err != nil {
            panic(err)
        }

        requests = append(requests, request)
    }

    return requests
}

func main() {
    // receive inputs
    payload := os.Getenv("PAYLOAD")
    if payload == "" {
        panic("Payload empty")
    }

    // parse payload
    requests := parsePayload(payload)

    // sync waiter for goroutines
    var wg sync.WaitGroup

    // make request
    for _, request := range requests {
        wg.Add(1)
        go func (request RequestSchema) {
            defer wg.Done()
            Request(request.Method, request.Url, request.Headers, request.Body)
        }(request)
    }

    // wait till all goroutines finish requesting
    wg.Wait()
}
