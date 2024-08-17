// Package restclient provides a simple HTTP client for making
// GET, POST, PUT, and DELETE requests and handling their responses.
package restclient

import (
    "bytes"
    "encoding/json"
    "errors"
    "io/ioutil"
    "net/http"
    "time"  // Import the time package
)

// RestClient defines the structure for making HTTP requests with
// a customizable timeout.
type RestClient struct {
    Client *http.Client
}

// NewRestClient initializes a new RestClient with a specified timeout.
//
// Example usage:
//
//     client := restclient.NewRestClient(10 * time.Second)
//
func NewRestClient(timeout time.Duration) *RestClient {
    return &RestClient{
        Client: &http.Client{Timeout: timeout},
    }
}

// Get sends a GET request to the specified URL and returns the response body as bytes.
// It returns an error if the request fails or if the response status code is not 200 (OK).
//
// Example usage:
//
//     body, err := client.Get("http://example.com/api/resource")
//     if err != nil {
//         log.Fatalf("Failed to make GET request: %v", err)
//     }
//     fmt.Println(string(body))
//
func (rc *RestClient) Get(url string) ([]byte, error) {
    resp, err := rc.Client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("failed to retrieve data from API, status code: " + resp.Status)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return body, nil
}

// Post sends a POST request with a JSON payload to the specified URL and returns
// the response body as bytes. It returns an error if the request fails or if
// the response status code is not 201 (Created).
//
// Example usage:
//
//     payload := map[string]string{"name": "example"}
//     body, err := client.Post("http://example.com/api/resource", payload)
//     if err != nil {
//         log.Fatalf("Failed to make POST request: %v", err)
//     }
//     fmt.Println(string(body))
//
func (rc *RestClient) Post(url string, payload interface{}) ([]byte, error) {
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    resp, err := rc.Client.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return nil, errors.New("failed to create resource, status code: " + resp.Status)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return body, nil
}

// Put sends a PUT request with a JSON payload to the specified URL and returns
// the response body as bytes. It returns an error if the request fails or if
// the response status code is not 200 (OK).
//
// Example usage:
//
//     payload := map[string]string{"name": "example"}
//     body, err := client.Put("http://example.com/api/resource/1", payload)
//     if err != nil {
//         log.Fatalf("Failed to make PUT request: %v", err)
//     }
//     fmt.Println(string(body))
//
func (rc *RestClient) Put(url string, payload interface{}) ([]byte, error) {
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := rc.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, errors.New("failed to update resource, status code: " + resp.Status)
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    return body, nil
}

// Delete sends a DELETE request to the specified URL and returns an error
// if the request fails or if the response status code is not 200 (OK).
//
// Example usage:
//
//     err := client.Delete("http://example.com/api/resource/1")
//     if err != nil {
//         log.Fatalf("Failed to make DELETE request: %v", err)
//     }
//
func (rc *RestClient) Delete(url string) error {
    req, err := http.NewRequest(http.MethodDelete, url, nil)
    if err != nil {
        return err
    }

    resp, err := rc.Client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return errors.New("failed to delete resource, status code: " + resp.Status)
    }

    return nil
}
