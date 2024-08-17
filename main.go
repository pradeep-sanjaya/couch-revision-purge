package main

import (
    "flag"
    "fmt"
    "github.com/pradeep-sanjaya/couch-revision-purge/config"
    "github.com/pradeep-sanjaya/couch-revision-purge/logger"
    "github.com/pradeep-sanjaya/couch-revision-purge/network"
    "github.com/pradeep-sanjaya/couch-revision-purge/couchdb"
    // "github.com/pradeep-sanjaya/couch-revision-purge/pulseapi"
    "log"
)

func main() {
    configFile := flag.String("config", "config.json", "Path to the configuration file")
    dbName := flag.String("dbname", "", "CouchDB database name")
    flag.Parse()

    if *dbName == "" {
        log.Fatalf("Database name is required")
        return
    }

    cfg, err := config.LoadConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load configuration: %v\n", err)
        return
    }

    if cfg.CIDR == "" || cfg.CouchDBPort == "" || cfg.APIEndpoint == "" {
        fmt.Println("Please provide a valid CIDR, CouchDB port, and API endpoint in the configuration file.")
        return
    }

    logger, err := logger.NewLogger(cfg.LogFile)
    if err != nil {
        log.Fatalf("Failed to open log file: %v\n", err)
    }

    // Use logger for all log output
    logger.Printf("Starting scan for CIDR: %s", cfg.CIDR)
    foundInstances := network.ScanNetwork(cfg.CIDR, cfg.CouchDBPort, logger, couchdb.IsCouchDBRunning)
    logger.Printf("Found %d CouchDB instances on the network.", foundInstances)

    foundIPs := network.ScanNetwork(cfg.CIDR, cfg.CouchDBPort, logger, couchdb.IsCouchDBRunning)

    if len(foundIPs) > 0 {
        for _, ip := range foundIPs {
            couchdbURL := fmt.Sprintf("http://%s:%s", ip, cfg.CouchDBPort)
            client := couchdb.NewCouchDBClient(couchdbURL, *dbName)

            // Example: Resetting a document by deleting all its revisions and recreating it
            err := client.ResetDocument(*dbName, logger)
            if err != nil {
                logger.Fatalf("Failed to reset document: %v", err)
            }

            // Check and delete the existing design document
            deleteMsg, err := client.CheckAndDeleteDesignDocument("rev_filter")
            if err != nil {
                logger.Fatalf("Failed to check and delete existing design document: %v", err)
            }
            logger.Println(deleteMsg)

            designDoc := map[string]interface{}{
                "views": map[string]interface{}{
                    "high_rev_gen": map[string]interface{}{
                        "map": "function(doc) { var revGen = parseInt(doc._rev.split(\"-\")[0]); if(revGen > 100000) { emit(doc._id, doc); } }",
                    },
                },
            }
            
            response, err := client.CreateDesignDocument("rev_filter", designDoc)
            if err != nil {
                logger.Fatalf("Failed to create design document: %v", err)
            }
            logger.Println("Design document created:", response)

            // Execute the GET request to query the design document
            queryResp, err := client.QueryDesignDocument("rev_filter")
            if err != nil {
                logger.Fatalf("Failed to query design document: %v", err)
            }
            logger.Println("Query result:", queryResp)

            // Handle the query response to delete conflicts
            err = client.HandleQueryResponse([]byte(queryResp))
            if err != nil {
                logger.Fatalf("Failed to handle query response: %v", err)
            }

            // Trigger database compaction
            compactResp, err := client.CompactDatabase()
            if err != nil {
                logger.Fatalf("Failed to compact database: %v", err)
            }
            logger.Println("Database compaction triggered:", compactResp)
        }
    } else {
        logger.Println("No CouchDB instances found.")
    }

    // expectedInstances, err := pulseapi.GetCouchDBInstanceCount(cfg.APIEndpoint)
    // if err != nil {
    //     logger.Fatalf("Failed to get CouchDB instance count from API: %v", err)
    // }
    // logger.Printf("API reports %d CouchDB instances.", expectedInstances)

    // if len(foundIPs) == expectedInstances {
    //     logger.Println("The number of CouchDB instances matches the API report.")
    // } else {
    //     logger.Printf("Mismatch: found %d instances, but API reports %d instances.", len(foundIPs), expectedInstances)
    // }

    logger.Println("Scan completed successfully.")
}