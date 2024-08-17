package pulseapi

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestGetCouchDBInstanceCount(t *testing.T) {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"couchdb_instances": 5}`))
    }))
    defer mockServer.Close()

    instances, err := GetCouchDBInstanceCount(mockServer.URL)
    if err != nil {
        t.Fatalf("Expected no error, got %v", err)
    }

    expectedInstances := 5
    if instances != expectedInstances {
        t.Errorf("Expected %d instances, got %d", expectedInstances, instances)
    }
}