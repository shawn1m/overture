#!/usr/bin/env bash

wget https://github.com/17mon/china_ip_list/raw/master/china_ip_list.txt
wget https://github.com/gfwlist/gfwlist/raw/master/gfwlist.txt
cp config.sample.json config.json

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o overture-darwin-amd64 main.go
GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -o overture-linux-386 main.go
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o overture-windows-amd64.exe main.go
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -o overture-windows-386.exe main.go

zip overture-darwin-amd64.zip overture-darwin-amd64 china_ip_list.txt gfwlist.txt config.json
zip overture-linux-386.zip overture-linux-386 china_ip_list.txt gfwlist.txt config.json
zip overture-windows-amd64.zip overture-windows-amd64.exe china_ip_list.txt gfwlist.txt config.json
zip overture-windows-386.zip overture-windows-386.exe china_ip_list.txt gfwlist.txt config.json
