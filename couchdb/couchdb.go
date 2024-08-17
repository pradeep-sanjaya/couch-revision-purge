package couchdb

import (
	"encoding/json"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/pradeep-sanjaya/couch-revision-purge/logger"
	"net"
	"time"
	"strings"
)

// CouchDBClient is a client for interacting with a CouchDB instance.
type CouchDBClient struct {
    BaseURL string
    DBName  string
}

// Document represents a document returned from a CouchDB query, including potential conflicts.
type Document struct {
    ID              string   `json:"_id"`
    Rev             string   `json:"_rev"`
    DeletedConflicts []string `json:"_deleted_conflicts,omitempty"`
}

// QueryResponse represents the structure of a CouchDB query response.
type QueryResponse struct {
    TotalRows int `json:"total_rows"`
    Offset    int `json:"offset"`
    Rows      []struct {
        ID    string   `json:"id"`
        Key   string   `json:"key"`
        Value Document `json:"value"`
    } `json:"rows"`
}

// IsCouchDBRunningFunc defines a function type that checks if CouchDB is running
// on a given IP address and port.
type IsCouchDBRunningFunc func(ip, port string) bool

// IsCouchDBRunning checks if CouchDB is running on the given IP address and port.
// It returns true if the service is reachable, and false otherwise.
//
// Example usage:
//
//     running := couchdb.IsCouchDBRunning("127.0.0.1", "5984")
//     if running {
//         fmt.Println("CouchDB is running on 127.0.0.1:5984")
//     }
//
func IsCouchDBRunning(ip, port string) bool {
    timeout := time.Second
    conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), timeout)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}

// NewCouchDBClient creates a new CouchDB client.
func NewCouchDBClient(baseURL, dbName string) *CouchDBClient {
    return &CouchDBClient{
        BaseURL: baseURL,
        DBName:  dbName,
    }
}

// GetDocument fetches a document by its ID.
func (c *CouchDBClient) GetDocument(docID string) (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/%s/%s", c.BaseURL, c.DBName, docID)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch document: %s", string(body))
    }

    var doc map[string]interface{}
    err = json.Unmarshal(body, &doc)
    if err != nil {
        return nil, err
    }

    return doc, nil
}

// GetAllRevisions fetches all revisions of a document by its ID.
func (c *CouchDBClient) GetAllRevisions(docID string) ([]string, error) {
    url := fmt.Sprintf("%s/%s/%s?revs_info=true", c.BaseURL, c.DBName, docID)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch document revisions: %s", string(body))
    }

    var doc struct {
        Revisions []struct {
            Rev string `json:"rev"`
        } `json:"_revs_info"`
    }
    err = json.Unmarshal(body, &doc)
    if err != nil {
        return nil, err
    }

    var revisions []string
    for _, revInfo := range doc.Revisions {
        revisions = append(revisions, revInfo.Rev)
    }

    return revisions, nil
}

// DeleteDocumentRevision deletes a specific document revision.
func (c *CouchDBClient) DeleteDocumentRevision(docID, rev string) (string, error) {
    url := fmt.Sprintf("%s/%s/%s?rev=%s", c.BaseURL, c.DBName, docID, rev)
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return "", err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("failed to delete document revision: %s", string(body))
    }

    return "Revision deleted successfully", nil
}

// DeleteAllRevisions deletes all revisions of a document by its ID.
func (c *CouchDBClient) DeleteAllRevisions(docID string, revisions []string) error {
    for _, rev := range revisions {
        resp, err := c.DeleteDocumentRevision(docID, rev)
        if err != nil {
            if strings.Contains(err.Error(), "not_found") {
                fmt.Printf("Revision %s is already deleted, skipping.\n", rev)
                continue
            }
            return fmt.Errorf("failed to delete revision %s: %v", rev, err)
        }
        fmt.Printf("Deleted revision %s: %s\n", rev, resp)
    }
    return nil
}

// DeleteDocument deletes a document by its ID.
func (c *CouchDBClient) DeleteDocument(docID string) error {
    url := fmt.Sprintf("%s/%s/%s", c.BaseURL, c.DBName, docID)
    req, err := http.NewRequest("DELETE", url, nil)
    if err != nil {
        return err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
        return fmt.Errorf("failed to delete document: %s", string(body))
    }

    return nil
}

// CreateDocument creates a new document.
func (c *CouchDBClient) CreateDocument(doc map[string]interface{}) error {
    url := fmt.Sprintf("%s/%s/%s", c.BaseURL, c.DBName, doc["_id"].(string))

    delete(doc, "_rev")

    jsonDoc, err := json.Marshal(doc)
    if err != nil {
        return err
    }

    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonDoc))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }

    if resp.StatusCode != http.StatusCreated {
        return fmt.Errorf("failed to create document: %s", string(body))
    }

    return nil
}

// ResetDocument resets a document by deleting all its revisions and recreating it.
func (c *CouchDBClient) ResetDocument(docID string, logger *logger.Logger) error {
    doc, err := c.GetDocument(docID)
    if err != nil {
        logger.Printf("Failed to fetch document: %v", err)
        return fmt.Errorf("failed to fetch document: %v", err)
    }

    revisions, err := c.GetAllRevisions(docID)
    if err != nil {
        logger.Printf("Failed to get revisions: %v", err)
        return fmt.Errorf("failed to get revisions: %v", err)
    }

    err = c.DeleteAllRevisions(docID, revisions)
    if err != nil {
        logger.Printf("Failed to delete all revisions: %v", err)
        return fmt.Errorf("failed to delete all revisions: %v", err)
    }

    err = c.DeleteDocument(docID)
    if err != nil {
        logger.Printf("Failed to delete document: %v", err)
        return fmt.Errorf("failed to delete document: %v", err)
    }

    err = c.CreateDocument(doc)
    if err != nil {
        logger.Printf("Failed to recreate document: %v", err)
        return fmt.Errorf("failed to recreate document: %v", err)
    }

    return nil
}

func (c *CouchDBClient) CompactDatabase() (string, error) {
    url := fmt.Sprintf("%s/%s/_compact", c.BaseURL, c.DBName)

    // Include an empty JSON body
    jsonBody := []byte(`{}`)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return "", err
    }

    req.Header.Set("Content-Type", "application/json") // Set Content-Type header

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    if resp.StatusCode != http.StatusAccepted {
        return "", fmt.Errorf("failed to trigger compaction: %s", string(body))
    }

    return string(body), nil
}

func (c *CouchDBClient) CheckAndDeleteDesignDocument(designDocName string) (string, error) {
    url := fmt.Sprintf("%s/%s/_design/%s", c.BaseURL, c.DBName, designDocName)

    // Fetch the design document to see if it exists
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusNotFound {
        return "Design document does not exist, no deletion needed", nil
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return "", fmt.Errorf("failed to fetch design document: %s", string(body))
    }

    // Parse the response to get the document's revision
    var doc struct {
        Rev string `json:"_rev"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
        return "", err
    }

    // Delete the existing design document
    deleteURL := fmt.Sprintf("%s?rev=%s", url, doc.Rev)
    req, err := http.NewRequest("DELETE", deleteURL, nil)
    if err != nil {
        return "", err
    }

    deleteResp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer deleteResp.Body.Close()

    if deleteResp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(deleteResp.Body)
        return "", fmt.Errorf("failed to delete design document: %s", string(body))
    }

    return "Existing design document deleted", nil
}

func (c *CouchDBClient) HandleQueryResponse(queryResponse []byte) error {
    var response QueryResponse
    err := json.Unmarshal(queryResponse, &response)
    if err != nil {
        return err
    }

    for _, row := range response.Rows {
        doc := row.Value
        if len(doc.DeletedConflicts) > 0 {
            fmt.Printf("Document %s has conflicts: %v\n", doc.ID, doc.DeletedConflicts)
            for _, conflictRev := range doc.DeletedConflicts {
                deleteResp, err := c.DeleteDocumentRevision(doc.ID, conflictRev)
                if err != nil {
                    return fmt.Errorf("failed to delete conflict for document %s: %v", doc.ID, err)
                }
                fmt.Printf("Deleted conflict revision %s for document %s: %s\n", conflictRev, doc.ID, deleteResp)
            }
        }
    }

    return nil
}

func (c *CouchDBClient) CreateDesignDocument(designDocName string, designDoc map[string]interface{}) (string, error) {
    url := fmt.Sprintf("%s/%s/_design/%s", c.BaseURL, c.DBName, designDocName)

    jsonDoc, err := json.Marshal(designDoc)
    if err != nil {
        return "", fmt.Errorf("failed to marshal design document: %v", err)
    }

    req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonDoc))
    if err != nil {
        return "", err
    }
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}

func (c *CouchDBClient) QueryDesignDocument(designDocName string) (string, error) {
    url := fmt.Sprintf("%s/%s/_design/%s/_view/high_rev_gen", c.BaseURL, c.DBName, designDocName)

    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}