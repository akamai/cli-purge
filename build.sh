#!/bin/bash
GOOS=darwin GOARCH=amd64 go build -o akamai-purge-mac-amd64 .
GOOS=linux GOARCH=amd64 go build -o akamai-purge-linux-amd64 .
GOOS=linux GOARCH=386 go build -o akamai-purge-linux-386 .
GOOS=windows GOARCH=386 go build -o akamai-purge-windows-386.exe .
GOOS=windows GOARCH=amd64 go build -o akamai-purge-windows-amd64.exe .

