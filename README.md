# Couch Revision Purge

Couch Revision Purge is a tool designed to scan a network for running CouchDB instances, retrieve the expected number of instances from an external API, and verify the match.

## Features

- Network scanning for CouchDB instances
- API interaction to retrieve CouchDB instance counts
- Logging of scan results and verification

## Installation

To install the project, clone the repository and build the project using Go:

```bash
git clone https://github.com/yourusername/couch-revision-purge.git
cd couch-revision-purge
go build -o couch-revision-purge
```

## Usage ##
```
./couch-revision-purge -config=config.json -dbname=parrott34974
go run main.go -config=config.json -dbname=parrott34974
```
