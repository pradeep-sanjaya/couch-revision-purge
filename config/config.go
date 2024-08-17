package config

import (
    "encoding/json"
    "os"
)

type Config struct {
    LogFile     string `json:"logfile"`
    CIDR        string `json:"cidr"`
    CouchDBPort string `json:"couchdbPort"`
    APIEndpoint string `json:"apiEndpoint"`
}

func LoadConfig(filename string) (*Config, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    decoder := json.NewDecoder(file)
    config := &Config{}
    err = decoder.Decode(config)
    if err != nil {
        return nil, err
    }

    return config, nil
}