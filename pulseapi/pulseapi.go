package pulseapi

import (
    "time"
    "encoding/json"
    "github.com/pradeep-sanjaya/couch-revision-purge/restclient"
)

type Response struct {
    CouchDBInstances int `json:"couchdb_instances"`
}

func GetCouchDBInstanceCount(apiURL string) (int, error) {
    client := restclient.NewRestClient(10 * time.Second)
    body, err := client.Get(apiURL)
    if err != nil {
        return 0, err
    }

    var apiResponse Response
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        return 0, err
    }

    return apiResponse.CouchDBInstances, nil
}