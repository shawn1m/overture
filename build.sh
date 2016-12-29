#!/usr/bin/env bash

mkdir -p release

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o release/overture-darwin-amd64 main.go
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o release/overture-linux-386 main.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o release/overture-windows-amd64.exe main.go
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o release/overture-windows-386.exe main.go

zip release/overture-darwin-amd64.zip release/overture-darwin-amd64
zip release/overture-linux-386.zip release/overture-linux-386
zip release/overture-windows-amd64.zip release/overture-windows-amd64.exe
zip release/overture-windows-386.zip release/overture-windows-386.exe