package restclient

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func TestRestClientGet(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"key": "value"}`))
    }))
    defer mockServer.Close()

    client := NewRestClient(10 * time.Second)
    body, err := client.Get(mockServer.URL)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    expectedBody := `{"key": "value"}`
    if string(body) != expectedBody {
        t.Errorf("Expected body %s, got %s", expectedBody, string(body))
    }
}