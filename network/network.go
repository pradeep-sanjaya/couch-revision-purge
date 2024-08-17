// Package network provides functions to scan a network for running
// CouchDB instances and return the results.
package network

import (
    "net"
    "sync"
    "github.com/pradeep-sanjaya/couch-revision-purge/couchdb"
    "github.com/pradeep-sanjaya/couch-revision-purge/logger"
)

// ScanNetwork scans all IPs in the provided CIDR network range for CouchDB instances.
// It uses goroutines to perform the scan concurrently and returns the number of
// instances found. The IsCouchDBRunning function is passed as a parameter to allow
// for mocking in tests.
func ScanNetwork(cidr string, couchDBPort string, logger *logger.Logger, isCouchDBRunning couchdb.IsCouchDBRunningFunc) []string {
    logger.Printf("Starting concurrent network scan on %s for CouchDB instances on port %s\n", cidr, couchDBPort)
    ips, err := Hosts(cidr)
    if err != nil {
        logger.Fatalf("Error parsing CIDR: %v\n", err)
    }

    var foundIPs []string
    var mu sync.Mutex
    var wg sync.WaitGroup

    for _, ip := range ips {
        wg.Add(1)
        go func(ip string) {
            logger.Printf("Scanning IP: %s\n", ip)
            defer wg.Done()
            if isCouchDBRunning(ip, couchDBPort) {
                logger.Printf("CouchDB running on IP: %s\n", ip)
                mu.Lock()
                foundIPs = append(foundIPs, ip)
                mu.Unlock()
            }
        }(ip)
    }

    wg.Wait()
    logger.Println("Network scan completed.")
    return foundIPs
}

// Hosts generates all possible IP addresses in the given CIDR range.
// It returns a slice of IP addresses as strings, excluding the network
// address and broadcast address.
//
// Example usage:
//
//     ips, err := Hosts("192.168.1.0/24")
//     if err != nil {
//         log.Fatalf("Failed to generate IPs: %v", err)
//     }
//     fmt.Println(ips)
//
func Hosts(cidr string) ([]string, error) {
    ip, ipnet, err := net.ParseCIDR(cidr)
    if err != nil {
        return nil, err
    }

    var ips []string
    for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
        ips = append(ips, ip.String())
    }

    return ips[1 : len(ips)-1], nil
}

// inc increments the IP address to iterate over all addresses in the range.
//
// Example usage:
//
//     ip := net.ParseIP("192.168.1.1")
//     inc(ip)
//     fmt.Println(ip) // 192.168.1.2
//
func inc(ip net.IP) {
    for j := len(ip) - 1; j >= 0; j-- {
        ip[j]++
        if ip[j] > 0 {
            break
        }
    }
}